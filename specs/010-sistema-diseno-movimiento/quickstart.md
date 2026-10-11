# Quickstart: F10 — Sistema de diseño y movimiento

**Feature**: F10 · **Branch**: `010-sistema-diseno-movimiento`

Cómo **levantar** el sitio público rediseñado y **validar** cada parte del alcance: sistema visual,
tema claro/oscuro (sistema → interruptor → persistencia), movimiento y movimiento reducido,
navegación fija sin scroll incómodo y selector de idioma compacto. Sirve como guía de QA manual
(SC-002, SC-007, SC-009, SC-012) y de verificación de **no-regresión** (SC-008).

> Nota: este documento se escribe en la fase de plan; los comandos y rutas son los del stack
> vigente (F1–F3). Si la implementación mueve algo, se actualiza aquí.

---

## 1. Levantar el stack

```bash
make up            # docker compose: PostgreSQL + backend + frontend en http://localhost:5173
make e2e           # Playwright local contra el stack real (frontend/e2e/)
```

Para iterar solo en frontend:

```bash
cd frontend
npm install
npm run dev        # http://localhost:5173
npm test -- --run  # unitarias (Vitest + Testing Library)
npm run typecheck && npm run lint
```

Apertura de referencia: la portada publicada en `/` (con contenido real cargado desde el panel, que
**no se ha tocado**). Útil también `/health` para comprobar que F1 sigue igual.

## 2. Validar el sistema visual (US1, SC-001/SC-002)

En **teléfono (320 px), tableta y computadora**, recorrer la portada completa (cabecera,
identidad/portada, «quiénes somos», horario, WhatsApp, contacto, redes, pie):

- [ ] Todas las secciones aplican el **mismo** sistema: tipografías de marca (Bebas Neue / Poppins /
      Playfair Display), proporción de color 70/20/10, espaciado y jerarquía coherentes.
- [ ] El conjunto se percibe **moderno y actual**, distinto del diseño previo y **no** como una
      plantilla genérica (comparación visual lado a lado; calibración con las referencias del
      cliente solo como inspiración — nada copiado).
- [ ] El **logotipo**: sin deformar, sin recolorear, sin girar, sin efectos, **en ambos temas y en
      todos los estados** (reposo, hover, foco, activo).
- [ ] Contenido real de distinta forma (textos largos, un servicio suelto, contenido solo en español
      navegando en inglés, varias redes): legible, sin huecos ni desbordes.
- [ ] Todo elemento interactivo se percibe como tal (hover, foco, activo) con contraste AA.
- [ ] Zoom al **200 %** y 320 px: sin desbordes ni pérdida de información.

## 3. Validar el tema claro/oscuro (US5, SC-012)

**Tres caminos** (en cada uno: contraste AA, logo con colores originales, mismo contenido y mismas
funciones que en claro):

1. **Sistema**: poner el dispositivo/ navegador en **modo oscuro** y abrir el sitio con una pestaña
   limpia (sin `ss.theme`) → el sitio se muestra **en oscuro**. (En Chrome DevTools:
   ⋮ → More tools → Rendering → *Emulate CSS prefers-color-scheme*.)
2. **Interruptor manual**: con el sitio abierto, usar el control de tema de la cabecera → el tema
   cambia **de inmediato, sin recargar**; **recargar** y **cerrar y reabrir** el navegador → la
   elección **persiste** (`localStorage['ss.theme']`).
3. **Primera visita sin elección**: borrar `ss.theme` (DevTools → Application → Local Storage) y
   alternar el modo del sistema → el sitio **sigue al sistema** (claro y oscuro). Si el sistema
   cambia **con el sitio abierto** y no hay elección manual, la interfaz lo sigue **sin recargar**.

Además:

- [ ] La elección manual **manda** hasta que se cambia (aunque el sistema cambie después).
- [ ] El control de tema es **compacto** (no ocupa espacio innecesario), accesible por teclado y con
      lector de pantalla (nombre claro del control y de su estado), foco visible.
- [ ] El sitio en oscuro **no muestra parpadeos** al cargar (ni «claro → oscuro» al abrir:
      anti-FOUC).
- [ ] 0 fallos de contraste AA en textos, botones, estados interactivos y foco **en ambos temas**
      (revisión por estado: reposo, hover, foco, activo, éxito/error si los hay).

## 4. Validar el movimiento y el movimiento reducido (US3, SC-005/SC-006)

**Con movimiento normal** (sistema sin «reducir movimiento»):

- [ ] Microinteraciones claras y agradables en hover, foco y activación; **sin desplazar** el maquetado.
- [ ] El contenido entra con transiciones **suaves y breves** (nivel moderado: nada llamativo ni
      largo) que no retrasan la lectura.
- [ ] Cambiar de idioma o de tema **no re-anima** secciones vistas ni provoca parpadeos ni
      descolocaciones.
- [ ] El contenido útil es visible e interactivo **desde el primer instante** (nunca bloqueado por
      una animación); fluidez en un dispositivo modesto.

**Con movimiento reducido** (sistema con «reducir movimiento»; DevTools → Rendering → *Emulate
prefers-reduced-motion: reduce*, o la opción del sistema):

- [ ] **0 animaciones** de entrada, revelado o desplazamiento: el contenido aparece **completo de
      inmediato**.
- [ ] **100 % de las funciones y la información** disponibles, idénticas al modo con movimiento.
- [ ] Activar «reducir movimiento» **con el sitio abierto** surte efecto sin recargar.

## 5. Validar la navegación (US2, SC-003/SC-004)

- [ ] Al abrir la portada, **todos** los enlaces de navegación están **a la vista** (sin menús
      ocultos, sin hamburguesa, sin desplegables, sin gestos).
- [ ] **Móvil (320 px)**: se llega a cualquier enlace **sin desplazar los enlaces de lado** y la
      página **nunca** provoca desplazamiento horizontal (comprobar también con ~8–10 enlaces si se
      simula la previsión de F4–F9).
- [ ] Al **bajar** por la página, la navegación **permanece visible y accesible** (barra fija), sin
      tapar contenido ni descolocar el maquetado.
- [ ] Clic en un enlace → la **sección destino y su título quedan visibles** (la barra no los tapa).
- [ ] Teclado: orden de foco comprensible, **foco visible siempre**, cada enlace indica a qué sección
      lleva; lector de pantalla: etiquetas correctas.
- [ ] Una persona con poca experiencia encuentra y usa la navegación con un solo clic o toque
      (prueba con personas, SC-003).

## 6. Validar el selector de idioma (US4, SC-007)

- [ ] **Menos espacio** que el actual (comparación directa en la misma vista: cabecera de F3 vs.
      cabecera de F10) conservando los textos completos **«Español»** y **«English»** (sin siglas,
      sin píldoras, sin menú, sin banderas).
- [ ] Identificable como selector de idioma por personas de poca experiencia y por lector de pantalla
      (nombre accesible = nombre completo del idioma).
- [ ] Comportamiento **idéntico a F3**: cambio inmediato; se mantiene al navegar; se recuerda entre
      visitas (`localStorage['ss.lang']`); primera visita sin preferencia → **español**.
- [ ] Igual en cabecera y en pie; sin desplazamientos ni tapados en ningún ancho.

## 7. Validar textos, accesibilidad y fronteras (SC-008/SC-009/SC-010)

- [ ] Cambiar a inglés: **0 textos de interfaz nuevos sin traducir** (incluidos los del control de
      tema); el contenido sin inglés muestra su fallback en español integrado en el diseño.
- [ ] Recorrido completo **por teclado y con lector de pantalla** en **ambos temas**: foco visible,
      orden comprensible, todo operable (referencia WCAG 2.1 AA; checklist manual de QA).
- [ ] **Panel de administración SIN CAMBIOS**: login, `/panel/informacion`, usuarios, roles y
      auditoría lucen y funcionan **exactamente igual que antes de F10** (F10B no existe aún);
      verificación con las suites e2e de panel y una pasada visual.
- [ ] `/health` sin cambios; rutas y URLs sin cambios.

## 8. Pruebas automatizadas (no-regresión incluida)

```bash
cd frontend
npm test -- --run        # unitarias: tema, movimiento, navegación, selector + TODAS las de F2/F3
npm run typecheck
npm run lint
make e2e                 # Playwright: diseno-movimiento, tema-oscuro, movimiento-reducido
                         # + portada-publica, portada-accesibilidad, portada-panel, portada-imagenes,
                         #   portada-patch-null, acceso, auditoria, status  (SIN CAMBIOS, SC-008)
make ci                  # pipeline completo
```

**Criterio de cierre**: todas las suites pasan **sin editar las pruebas existentes** (SC-008). Si una
prueba de F3 exigiera un cambio visual que la propuesta aprobada provoca, se escala al humano: no se
modifica la prueba para hacerla pasar.

## 9. Revisión de aceptación (humano)

- [ ] La propuesta visual elegida (una de las 3 de `ux.md`, FR-026) se reconoce en la implementación
      y el cliente la califica como **moderna y acorde a su marca** («no parece un diseño típico y
      común de IA») (SC-002).
- [ ] 0 elementos copiados de las referencias (`iglesiaoasis.com`, `river-arena.org`) (SC-013).
- [ ] Cimientos para F4–F9 entregados y documentados (`ux.md` + este plan): una sección nueva
      construida con ellos encaja sin rediseño (SC-011).
