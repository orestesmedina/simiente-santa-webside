# Constitución: principios del kit

Principios no negociables para cualquier proyecto, sea cual sea su tecnología. Todos los agentes deben cumplirlos.
Niveles: **DEBE** (obligatorio), **DEBERÍA** (recomendado), **PUEDE** (opcional).

## I. La especificación manda
- La spec aprobada es la fuente de verdad. El código DEBE implementar la spec, no interpretarla.
- Todo cambio de comportamiento DEBE reflejarse primero en `spec.md`.
- Cada requisito DEBE tener criterios de aceptación verificables.

## II. Arquitectura
- La arquitectura es del proyecto: sus partes y tecnologías están en el perfil (`equipo/perfil.json`), y sus reglas, en las reglas del proyecto.
- Cuando dos partes del sistema se comunican, su contrato DEBE escribirse antes que el código.
- Las dependencias externas DEBERÍAN minimizarse; toda dependencia nueva requiere justificación en `plan.md`.

## III. Pruebas
- Toda tarea de implementación DEBE incluir sus pruebas, salvo que las reglas del proyecto digan otro alcance (qué se prueba y qué no).
- Ningún cambio se integra con pruebas fallando.
- La cobertura mínima y los tipos de prueba obligatorios los fijan las reglas del proyecto.

## IV. Seguridad
- DEBE prevenirse la inyección: solo consultas parametrizadas, nunca concatenación de texto. (CWE-89)
- DEBE validarse toda entrada en el servidor, aunque el cliente ya valide. (CWE-20)
- Todo dato que se muestre DEBE escaparse según dónde se muestra. (CWE-79)
- Las contraseñas DEBEN guardarse con un algoritmo de hash hecho para contraseñas (bcrypt o argon2). Nunca en texto plano. (CWE-256)
- Los secretos DEBEN venir de variables de entorno; nunca en el repositorio. (CWE-798)
- Todo acceso protegido DEBE verificar autenticación y autorización en el servidor. (CWE-862)
- Las dependencias DEBEN pasar la auditoría del proyecto (verbo `auditar`) sin vulnerabilidades altas o críticas.

## V. Calidad de código
- El código DEBE pasar sin errores el formato y la revisión del proyecto (verbos `formato` y `revisar`).
- Funciones cortas y con una sola responsabilidad. Nombres descriptivos.

## VI. Base de datos
- Todo cambio de esquema DEBE quedar en un archivo versionado, con la forma de revertirlo.
- Nunca se modifica un cambio de esquema ya aplicado; se crea uno nuevo.

## VII. Operación
- La configuración DEBE venir de variables de entorno.
- El proyecto DEBE poder levantarse en local con pasos documentados y repetibles.

## VIII. Gobierno
- La spec y el plan requieren aprobación humana antes de implementar.
- Los despliegues a producción requieren aprobación humana.
- Las reglas del proyecto pueden agregar o endurecer estos principios. Solo pueden reducirlos donde el principio lo permite expresamente.
- En un proyecto existente, lo que hoy no cumple un principio se registra como deuda; el código nuevo DEBE cumplirlo.
- Los principios y las reglas del proyecto solo se modifican con aprobación de la dirección técnica.

**Versión de los principios:** 2.0.0
