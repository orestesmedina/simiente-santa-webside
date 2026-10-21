---
name: convenciones-proyecto
description: Cómo redactar las convenciones propias de un proyecto (nombres, estructura, patrones de diseño, bibliotecas) leyendo su código, y dejarlas como una skill del proyecto. Usar con el comando bowser-conventions, al instalar el kit en un proyecto que ya tiene código, o cuando el proyecto cambia su forma de hacer algo.
---
# Las convenciones del proyecto

Las skills del catálogo del kit dicen **el estándar** de cada tecnología: lo que recomienda su documentación oficial. No dicen cómo lo hace *este* proyecto: cómo nombra sus variables, sus tablas y sus columnas, qué patrones de diseño sigue, qué bibliotecas eligió. Eso son sus convenciones, y se escriben aquí: una skill por parte del proyecto, que es del proyecto y vive en su `.agents/skills/`.

Sirven para que el código nuevo se parezca al que ya hay, y para saber en qué se aparta el proyecto del estándar.

## Qué manda
1. **Seguridad: siempre el estándar.** Una convención del proyecto nunca autoriza a saltarse una regla de seguridad de la skill del catálogo.
2. **Nombres, estructura, patrones y bibliotecas: el proyecto.**
3. **Lo que el proyecto no define: el estándar.**

## Formato
Un archivo por parte: `.agents/skills/convenciones-de-<parte>/SKILL.md`, con el nombre que la parte tiene en el perfil.

```markdown
---
name: convenciones-de-api
description: Convenciones propias de la parte api de este proyecto (nombres, estructura, patrones, bibliotecas). Usar siempre que se cree o modifique código en api/.
---
# Convenciones de api

Redactadas el <fecha> leyendo el código. Estándar de referencia: skill `<la del catálogo>`.

## Nombres
- <regla, en una línea>. Ej.: `<fragmento real>` (`<archivo>`).

## Estructura
- …

## Patrones
- …

## Base de datos
- … (solo si la parte tiene `dev-dba` junto a otro rol)

## Bibliotecas y herramientas
- …

## Dónde se aparta del estándar
| El estándar dice | Aquí se hace | En código nuevo |
|---|---|---|
| <regla de la skill del catálogo> | <lo que hace el proyecto, con su archivo> | Se sigue el proyecto |

## Deuda
- <qué pide el estándar, dónde no se cumple>
```

## Cómo redactarlas
1. **Lee antes de escribir.** Recorre el código propio de la parte (no el de terceros que declara el perfil): archivos de configuración, varios archivos de cada carpeta, los más recientes y los más grandes. En una base de datos, el esquema y varias consultas o procedimientos.
2. **Describe lo que el código hace, no lo que debería hacer.** Una convención es algo que el proyecto repite. Si lo viste una sola vez, no es una convención.
3. **Cada regla lleva un ejemplo real** y el archivo de donde salió. Sin ejemplo, no se escribe: es una suposición.
4. **Donde el código no es consistente** (dos formas de hacer lo mismo), no elijas tú: muestra las dos, di cuál es más frecuente y más reciente, y pregunta a la persona cuál vale para lo nuevo.
5. **Compara con el estándar.** Lee la skill del catálogo de esa tecnología y llena "Dónde se aparta del estándar". Para cada diferencia, la columna "En código nuevo" dice qué se hace:
   - Si es de forma (nombres, estructura, estilo, patrones, bibliotecas): se sigue el proyecto, salvo que la persona diga que lo nuevo adopta el estándar.
   - **Si es de seguridad: no se escribe como convención.** Va a "Deuda", con dónde ocurre y qué pide el estándar. Decidir cuándo y cómo se corrige es de una persona.
6. **No copies el estándar.** Lo que el proyecto hace igual que la skill del catálogo no se repite aquí: esta página solo dice lo propio.
7. Qué cubrir, según la parte:
   - **Código:** cómo se nombran archivos, clases, funciones y variables; idioma de los nombres y de los comentarios; sangría y formato; cómo se organizan las carpetas y dónde va cada tipo de archivo; cómo se cargan las dependencias; los patrones que se repiten (capas, acceso a datos, manejo de errores, validación, respuestas); las bibliotecas que usa y en qué versión; dónde y cómo están las pruebas.
   - **Base de datos:** cómo se nombran tablas, columnas, claves, índices y procedimientos; prefijos; tipos que usa para identificadores, fechas, dinero y estados; cómo se hacen los cambios de esquema y dónde quedan; si la lógica vive en la base de datos o en el código.
   - **Interfaz:** biblioteca de componentes o plantilla visual; cómo se arma una página; cómo se nombran las clases de estilo; cómo se llama al servidor; cómo se muestran los errores y las cargas.
8. **Corta y concreta.** Una regla, un ejemplo: no varias reglas ni varios ejemplos en la misma línea. El archivo se carga en cada tarea, así que el límite es de tamaño, no de líneas: **unos 7 000 caracteres** (compruébalo con `wc -c`). Si pasa, sobra detalle: quita lo que el proyecto hace igual que el estándar, las excepciones heredadas que nadie va a repetir y los segundos ejemplos.
9. **Si la parte tiene `dev-dba` junto a otro rol**, lo de la base de datos (nombres de tablas, columnas e índices, tipos, cambios de esquema, consultas) va en una sección propia, `## Base de datos`, después de "Patrones": es lo único del archivo que ese rol necesita, y así lo encuentra sin leer el resto.
10. Sin secretos ni datos reales en los ejemplos.

## En un proyecto nuevo
No hay código que leer. Las convenciones nacen de lo que decide el plan (qué bibliotecas, qué estructura) y se escriben cuando la primera funcionalidad ya dejó código: antes serían un deseo, no una convención.

## Entrega
- El archivo propuesto de cada parte.
- Un resumen en lenguaje simple: lo más característico del proyecto, las diferencias con el estándar, la deuda de seguridad encontrada y las preguntas abiertas.

Con la aprobación de la persona, el archivo se escribe y la skill se agrega a las `skills` de esa parte en el perfil (`equipo/perfil.json`), junto a la del catálogo. Cambiar el perfil se confirma como siempre: lo hace una persona.
