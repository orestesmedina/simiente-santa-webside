# Entrega F7 — Ofrendas y diezmos

**Fecha:** 2026-10-16 · **Versión:** 0.7.0 · **Rama:** `007-donaciones`

Nota de entrega en lenguaje sencillo. Si querés la parte técnica, está en el [`README`](../../README.md), en el [`CHANGELOG`](../../CHANGELOG.md) y en la [spec de la funcionalidad](../../specs/007-donaciones/spec.md).

## Qué se entrega

Ahora el sitio **muestra cómo apoyar a la iglesia** y el equipo lo administra solo, desde el panel. Nada de esto existía antes.

- **La sección «Ofrendas y diezmos»** (en <http://localhost:5173/ofrendas>), visible **sin cuenta ni registro**: primero un **texto de introducción** («cómo donar»), después las **cuentas bancarias (IBAN)** y los **números de SINPE Móvil**, y al final el **texto que explica en qué se usan las donaciones**. Cada cuenta o número muestra su **tipo**, el **número**, el **banco** y —si ustedes lo cargan— el **titular**; en las cuentas bancarias también se ve si es en **colones o en dólares** (el SINPE Móvil no lleva moneda).
- **Los números se leen y se copian sin errores**: el IBAN se muestra **agrupado en bloques** para leerlo mejor, pero un botón **«Copiar número»** copia el **número exacto** (agrupado es solo la presentación). El SINPE se muestra en su **formato normalizado**.
- **Los números se validan al guardarlos**: el **IBAN** se revisa por estructura y longitud (costarricenses e internacionales válidos) y el **SINPE Móvil** tiene que ser un **celular costarricense de 8 dígitos** (se acepta con o sin `+506`). Lo que no cumpla se rechaza con un aviso claro y **no se guarda nada a medias**.
- **El orden lo deciden ustedes**: cada cuenta y cada número tiene un número de orden y con él se reordenan; las cuentas bancarias van en un grupo y los SINPE en otro.
- **En español y en inglés**, igual que el resto del sitio: lo que no esté traducido se muestra en español (nunca un hueco). Eso sí, los **números y datos de la cuenta no se traducen** (son siempre iguales).
- **Todo es opcional menos lo básico**: de cada cuenta son obligatorios el **tipo**, el **número** y el **banco**; el **titular es opcional** (si no lo cargan, simplemente no se muestra esa línea).
- **Aparece solo cuando hay contenido**: mientras no haya nada publicado, la entrada **«Ofrendas y diezmos»** no se muestra en la barra, el menú ni el pie, **y si alguien abre la dirección directamente lo manda a la portada** (sin avisos ni pantallas raras). Es igual que hacemos en Ministerios.
- **El bloque «Ofrendas y diezmos» en la portada**, colocado **justo debajo del bloque de Ministerios**: hasta **5 formas de ofrendar** (las primeras según el orden que ustedes fijen) con el botón **«Ver formas de ofrendar»** para llegar a la sección completa. **Si no hay ninguna cuenta publicada, el bloque no aparece** (la portada queda limpia).
- **El panel para administrarla**: **`/panel/donaciones`**, con alta, edición, publicación, retirada y eliminación con confirmación; listado agrupado por tipo con reordenación; y la edición de los **dos textos** (introducción y uso). En Inicio aparece el acceso directo **«Gestión de donaciones»**.
- **Un permiso existente, sin permisos nuevos**: se usa **«Donaciones»**, el permiso que ya estaba reservado desde la gestión de usuarios (F2). Quien no lo tiene, **no ve ni puede tocar el módulo** —el servidor lo comprueba en cada operación y responde con un mensaje claro—. El catálogo de permisos **sigue teniendo 10**.
- **Todo queda registrado**: cada alta, edición, publicación, retirada, edición de texto, eliminación y denegación aparece en `/panel/auditoria` con texto legible, igual que en F2, F3, F4 y F6.
- **Se ve bien en cualquier dispositivo**: teléfono, tableta y computadora, con letras legibles y botones grandes, pensando en personas que recién se inician en internet.

## Lo importante: publicar un número de cuenta es delicado

Como aquí un error de lectura puede hacer que un aporte no llegue, el sistema pone el foco en la **exactitud**:

- **El número que se publica es exactamente el que ustedes guardaron** (no se recorta, no se cambia). La agrupación del IBAN es solo para leerlo mejor; al copiar, se copia el número completo.
- **Un cambio guardado se ve de inmediato** en el sitio: si el banco cambia la cuenta, la corrigen y el dato viejo deja de mostrarse en el acto.
- **Nada a medio hacer**: si un dato está mal, no se guarda ninguna parte; se avisa qué corregir.

## Qué queda fuera por ahora (y por qué)

- **El sitio no cobra ni procesa donaciones**: no hay pagos con tarjeta, montos, metas ni comprobantes. Muestra la información (las cuentas y los números) y **la persona dona por sus propios medios**.
- **No se registra quién donó**: no hay historial de donaciones ni recibos.
- **No se verifica** que la cuenta exista, siga activa o sea de la iglesia: solo se revisa que el número esté bien escrito. La corrección es manual, editando el dato.
- **Sin imágenes** en esta sección (texto y datos) y **sin formulario** para inscribirse o avisar de una donación.

## Limitación conocida (interna, no afecta al sitio)

Hay un **defecto preexistente en las pruebas automáticas** del proyecto (no en el producto): una de las pruebas de base de datos puede fallar de forma intermitente en el servidor de integración continua. Es anterior a esta entrega, está documentado en el **defecto conocido 23** del [roadmap](../../docs/producto/roadmap.md) (con su causa y el arreglo propuesto) y **no afecta a lo que ven los visitantes ni el equipo**.

## Pruebas que se harán con el sitio ya en producción

Dos comprobaciones **no se pueden automatizar** y quedan para después del despliegue, igual que en funcionalidades anteriores: la **prueba de usabilidad** con personas de distintas edades (que al menos 9 de cada 10 encuentren cómo donar sin ayuda, y que lo logren en menos de 30 segundos) y la **revisión con un lector de pantalla real**. El protocolo completo está en [`specs/007-donaciones/pruebas-usabilidad-SC-012.md`](../../specs/007-donaciones/pruebas-usabilidad-SC-012.md).

## Cómo probarlo en dos minutos

1. Levantá el sitio (`make up`) y entrá al panel con una cuenta que tenga el permiso **«Donaciones»** (o con el administrador inicial).
2. En **`/panel/donaciones`**, creá una cuenta IBAN (tipo, número, banco obligatorios; titular y moneda —colones o dólares— opcionales) y publicala; agregá un número de SINPE Móvil (8 dígitos, con o sin `+506`).
3. Escribí el **texto de introducción** y el de **en qué se usan las donaciones**, y publicalos.
4. Abrí <http://localhost:5173/ofrendas> y comprobá que se ve todo; probá el botón **«Copiar número»** y cambiá de idioma.
5. Volvé a la portada (<http://localhost:5173/>) y verás el bloque **debajo del de Ministerios**.
6. Retirá todo a borrador y abrí `/ofrendas` de nuevo: te manda a la portada, sin avisos, y la entrada del menú desaparece.
