---
name: postgres-db
description: Convenciones para diseñar tablas, índices, consultas, cambios de esquema y funciones en PostgreSQL, según el manual oficial y la lista "Don't Do This" de la wiki de PostgreSQL. Usar al crear o modificar tablas, índices, consultas SQL o funciones.
metadata:
  madurez: redactada
---
# PostgreSQL — convenciones

Fuentes: el manual de PostgreSQL (postgresql.org/docs/current/) y la página "Don't Do This" de su wiki (wiki.postgresql.org/wiki/Don't_Do_This). Revisado el 2026-10-09. Ante una duda que esta página no resuelva, se consulta ahí. Donde el manual no fija una convención (el formato de los archivos de cambio), esta página lo dice.

## Versión
- Cada versión mayor tiene soporte durante cinco años. Con soporte al 2026-10-09 (postgresql.org/support/versioning): de la 14 a la 18; la 14 termina el 2026-11-12.
- Un proyecto usa una versión con soporte y se mantiene en su última versión menor. Si corre en una sin soporte, dilo en la entrega como riesgo.

## Nombres
- **En minúsculas y con guion bajo** (`order_items`), sin comillas dobles. Un nombre con mayúsculas obliga a escribirlo entre comillas para siempre.
- Sin palabras reservadas (`user`, `order`, `group`). Máximo 63 caracteres.
- En inglés, tablas en plural. La clave primaria se llama `id` y una clave foránea, `<tabla en singular>_id`.

## Tipos de datos
- Texto: `text`. No `char(n)`, y `varchar(n)` solo si el límite es una regla real del negocio; un límite de longitud se expresa mejor con un `CHECK`.
- Fechas con hora: **`timestamptz`** siempre. `timestamp` sin zona no representa un instante.
- Dinero y todo valor exacto: `numeric`. Nunca el tipo `money`, ni `real` o `double precision`.
- Claves que se generan solas: columnas de identidad (`bigint GENERATED ALWAYS AS IDENTITY`), no `serial`. Si hace falta un identificador que no se pueda adivinar, `uuid` con `gen_random_uuid()`.
- Verdadero o falso: `boolean`. Documentos: `jsonb`, no `json`.
- La base de datos se crea en UTF-8, nunca `SQL_ASCII`.

## Tablas
- Clave primaria en toda tabla.
- Claves foráneas declaradas, con el `ON DELETE` decidido a conciencia. PostgreSQL **no** crea el índice de la columna que referencia: se crea a mano.
- `NOT NULL` por defecto; una columna acepta nulos solo si el negocio lo necesita.
- Las reglas de los datos se declaran en la tabla (`UNIQUE`, `CHECK`, claves foráneas), no solo en el código.
- No se usan reglas (`CREATE RULE`) ni herencia de tablas: en su lugar, disparadores y particiones.

## Índices
- Índice en las columnas de los `WHERE`, `JOIN` y `ORDER BY` frecuentes, y en cada clave foránea. En un índice de varias columnas, primero las que se comparan por igualdad.
- Cada índice cuesta en cada escritura: no se crea uno que ninguna consulta usa.
- En una tabla con datos, `CREATE INDEX CONCURRENTLY`, que no bloquea las escrituras. No puede ir dentro de una transacción: va solo, en su propio archivo de cambio.
- Toda consulta nueva de listado o reporte se revisa con `EXPLAIN (ANALYZE, BUFFERS)`.

## Consultas
- Siempre con parámetros (`$1`, `$2`); nunca uniendo texto.
- Se listan las columnas que se necesitan; nada de `SELECT *`.
- `NOT EXISTS` en vez de `NOT IN`: con un nulo en la lista, `NOT IN` no devuelve nada.
- Rangos de fechas con `>=` y `<`, no con `BETWEEN`, que incluye los dos extremos.
- Lo que modifica varias tablas va en una transacción.
- Listados grandes: paginación por la clave (`WHERE id > $1 ORDER BY id LIMIT $2`), no con un `OFFSET` alto.

## Cambios de esquema
- Todo cambio va en un **archivo nuevo**, numerado en orden, con su pareja que lo revierte. PostgreSQL no trae una herramienta para esto: se usa la del proyecto, con su formato y su carpeta. Si no tiene ninguna, carpeta `migrations/` en la parte de la base de datos, con `0007_add_orders.up.sql` y `0007_add_orders.down.sql`.
- Un archivo de cambio ya versionado no se edita nunca: se corrige con otro.
- En PostgreSQL los cambios de estructura **sí** van dentro de una transacción: un archivo de cambio se aplica entero o no se aplica.
- Un `ALTER TABLE` toma un bloqueo que detiene las lecturas y escrituras mientras dura. En una tabla con datos:
  - Fijar `lock_timeout` antes, para que falle en vez de dejar a todos esperando.
  - Una restricción nueva se agrega con `NOT VALID` y se valida después, con `VALIDATE CONSTRAINT`, que no bloquea las escrituras.
  - Una columna obligatoria nueva, en pasos: agregarla aceptando nulos, rellenarla por partes, y recién entonces restringir.
- Un `down` que borra una columna o una tabla pierde datos: dilo antes de escribirlo.
- Nada de datos reales, usuarios ni contraseñas en un archivo de cambio.

## Funciones y procedimientos
- Solo si el plan los justifica. Nombres en minúsculas, como las tablas; parámetros y variables con prefijo (`p_`, `v_`) para que no se confundan con columnas.
- **Nada de SQL dinámico armado con texto recibido.** Si hace falta `EXECUTE`, los valores van con `USING` y los nombres con `format('%I', …)`.
- Una función `SECURITY DEFINER` se ejecuta con los permisos de quien la creó: solo si es imprescindible, y siempre fijando `search_path` en su definición, como indica el manual.
- Cambiar los parámetros o lo que devuelve cambia el contrato con el código que la llama: va en la entrega, con el antes y el después.

## Seguridad
- La aplicación se conecta con un rol propio, sin `SUPERUSER`, con solo los permisos que usa.
- Contraseñas de conexión con `scram-sha-256`. Nunca autenticación `trust` por red.
- Ninguna credencial en los archivos del repositorio.

## Cómo probar un cambio
- Siempre contra una base de datos local o de pruebas, en la misma versión que producción. Nunca contra una con datos reales.
- La secuencia completa: aplicar el `up`, comprobar el resultado, aplicar el `down`, comprobar que quedó como estaba, y aplicar el `up` otra vez.
- Si el proyecto tiene pruebas de base de datos, las nuevas van en su misma carpeta; si no, las consultas de comprobación y su resultado esperado van en la entrega, para que QA las repita.

## Lo que no va en git
- Los volcados con datos reales (los que produce `pg_dump` sobre una base en uso) y la carpeta de datos del motor, si la base local corre en un contenedor con una carpeta del proyecto.
- El esquema, los archivos de cambio y los datos de ejemplo inventados sí van en git.

## Si el proyecto tiene sus propias convenciones
- Esta página es el estándar. Si el proyecto tiene una skill de convenciones propias, **en nombres, estructura, patrones y herramientas manda esa**, para que lo nuevo se parezca a lo que ya hay. Lo que ella no diga se hace como dice esta página.
- **Seguridad**, los parámetros en las consultas y el archivo de cambio con su reverso no admiten excepción. Si cumplirlos exige cambiar algo que tu tarea no cubre, detente y avisa.
- No renombres ni reorganices lo que tu tarea no toca. Lo que encuentres fuera del estándar va en la entrega, como deuda.
