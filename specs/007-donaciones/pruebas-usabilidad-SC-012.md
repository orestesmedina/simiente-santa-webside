# Protocolo de prueba de usabilidad SC-012 (y SC-001) — F7 Ofrendas y diezmos

Criterio **SC-012**: «En pruebas con personas de distintas edades, incluidas con poca experiencia en internet, al menos el **90 %** identifica **cómo puede donar** (a qué cuenta o número) y **en qué se usan las donaciones**, sin ayuda.» Incluye el **cronómetro de los 30 s** de **SC-001** («encuentra cómo donar en menos de 30 segundos desde que abre la sección»). Complementa **SC-008** (revisión con **lector de pantalla real**).

**Estado: PENDIENTE** (no es automatizable; se ejecuta al cerrar la fase de validación / tras el despliegue). Responsable: el equipo (o la persona que designe el humano). No bloquea el APROBADO de QA, pero **sí el cierre de la funcionalidad**, igual que en F4 (SC-010) y F6 (SC-012).

La prueba es con el producto terminado y real, no de laboratorio:

- Con el stack levantado (`make up`) y contenido **real publicable** por el equipo (no «Lorem ipsum»): al menos **dos cuentas IBAN** (una en colones y una en dólares) y **dos números de SINPE Móvil**, con y sin titular (para comprobar que la ficha sin titular se ve completa), y los dos textos (introducción y uso) publicados. Dejar además **un método en borrador** que **no** debe verse y, si se puede, una segunda ronda con **nada publicado** (para comprobar la redirección).
- En el dispositivo que la persona use a diario y, si se puede, una segunda ronda en el otro extremo (computadora o teléfono).
- En español; el inglés solo si la persona quiere cambiarlo por su cuenta (no se fuerza).

## Muestra

- **≥ 6 personas**, distintas edades («adultos mucho más mayores que saben lo básico de internet» incluidos, `idea.md` §7), mixto de experiencia con internet.
- Idealmente 1–2 que nunca hayan usado el panel (solo visitantes).

## Guion de la sesión (~10 min por persona)

1. Dejar abierta la portada del sitio real. **Sin dar contexto**: «Bienvenido/a; te pido que uses el sitio como lo harías normalmente».
2. Tareas en este orden (sin decir dónde está nada):
   1. «¿Cómo haría la iglesia para recibir una ayuda económica tuya? ¿Dónde lo buscarías?» — solo cuenta como éxito si llega **por su cuenta** (por el menú, por el bloque de la portada o por el pie).
   2. «¿A qué cuenta o número tendrías que enviar el aporte? Léemelo / cópialo.» — aquí corre el **cronómetro de SC-001**: desde que abre la sección hasta que identifica el número. Se considera **exacto** si el número que lee (o copia) coincide **carácter por carácter** con el publicado. Si se frustra, detener a los 60 s y anotar «no alcanzado».
   3. «¿Y en qué se usan las donaciones de la iglesia?» — debe leerlo en la sección (texto de uso).
   4. «Si el banco les cambiara la cuenta, ¿dónde crees que se corregiría?» — se acepta cualquier respuesta que apunte al panel (no se exige que lo sepan).
3. Con una cuenta **sin titular** publicada: comprobar que la persona **no** percibe huecos ni datos «en blanco».
4. Si la persona usa **lector de pantalla** o navegación por teclado: recorrer `/ofrendas` completo y activar **«Copiar número»** (esto cubre **SC-008**).

## Qué se registra

Por persona: si encontró la sección por su cuenta (sí/no), el **tiempo** hasta identificar el número, si el número leído/copiado fue **exacto** (sí/no), si identificó **en qué se usan** las donaciones (sí/no), el dispositivo usado y cualquier comentario. Se cierran las dos rondas del cronómetro con los números reales.

## Criterio de aceptación (SC-012 / SC-001 / SC-008)

- **≥ 90 %** de las personas (al menos 6 de cada 6 personas si son 6) identifica **cómo donar** y **en qué se usan** las donaciones **sin ayuda**.
- El tiempo mediano hasta identificar el número es **< 30 s** (SC-001), y **0** números leídos o copiados con error (SC-009 en la práctica).
- **0 hallazgos críticos** de accesibilidad con el lector de pantalla real (SC-008) y recorrido completo por teclado.
- Si un resultado no se cumple, se abre un hallazgo y se corrige por el flujo normal (bug o mejora), igual que en F4/F6.
