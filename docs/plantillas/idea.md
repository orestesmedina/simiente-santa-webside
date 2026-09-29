# Idea: [nombre del producto]

<!--
Plantilla del kit. Cópiala a docs/producto/idea.md y complétala TÚ, sin IA:
es la materia prima de todo lo demás. Si no puedes llenar una sección, esa es
una pregunta para el cliente, no para el agente.

Extensión ideal: una página. Frases cortas. Nada de tecnología (lenguajes,
frameworks, base de datos): eso lo decide el arquitecto después.
Borra estos comentarios al terminar.
-->

**Autor:** [nombre] · **Cliente / interesado:** [nombre o "interno"] · **Fecha:** [AAAA-MM-DD]

## 1. Problema

<!-- ¿Qué duele hoy, a quién y cuánto? Una o dos frases. Describe el problema, no la solución. -->

[Ej.: El equipo de soporte pierde solicitudes porque llegan por WhatsApp, correo y llamadas, y nadie sabe cuáles están pendientes.]

## 2. Usuarios

<!-- Quién usará el sistema. Un renglón por tipo de usuario, con lo que necesita hacer. -->

| Usuario | Qué necesita hacer |
|---|---|
| [Ej.: Agente de soporte] | [Registrar solicitudes y ver las que tiene asignadas] |
| [Ej.: Supervisor] | [Asignar solicitudes y ver cuántas hay pendientes por persona] |

## 3. Cómo se resuelve hoy

<!-- Qué usan actualmente (Excel, papel, otro sistema) y qué es lo que falla. Ayuda a no repetir el problema. -->

[Ej.: Una hoja de Excel compartida que nadie actualiza a tiempo.]

## 4. Qué sería un éxito

<!-- Resultados medibles. Evita "rápido", "fácil" o "moderno" sin número. -->

- [Ej.: Ninguna solicitud queda más de 24 horas sin asignar.]
- [Ej.: El supervisor ve el estado de todas las solicitudes en una sola pantalla.]

## 5. Lo mínimo que debe hacer (MVP)

<!-- Lista corta de capacidades imprescindibles para resolver el problema principal. Si dudas de algo, va a la sección 6. -->

1. [Ej.: Registrar una solicitud con título, descripción y cliente.]
2. [Ej.: Asignarla a un agente.]
3. [Ej.: Cambiar su estado: pendiente, en proceso, resuelta.]

## 6. Fuera de alcance (por ahora)

<!-- Lo que explícitamente NO entra en esta versión. Evita discusiones después. -->

- [Ej.: Integración con WhatsApp.]
- [Ej.: Aplicación móvil.]

## 7. Restricciones

<!-- Todo lo que limita la solución. Deja "ninguna conocida" si aplica, pero no lo dejes vacío. -->

- **Fecha límite:** [Ej.: primera versión usable el 15 de noviembre]
- **Integraciones obligatorias:** [Ej.: ninguna / facturación electrónica / pagos]
- **Datos sensibles:** [Ej.: nombres y teléfonos de clientes → requiere control de acceso]
- **Usuarios y volumen esperado:** [Ej.: 10 agentes, ~200 solicitudes por día]
- **Otras:** [Ej.: debe funcionar en el celular del supervisor]

## 8. Preguntas abiertas

<!-- Lo que todavía no sabes. Es mejor listarlo aquí que dejar que el agente lo invente. -->

- [Ej.: ¿Un cliente puede ver el estado de su propia solicitud?]
