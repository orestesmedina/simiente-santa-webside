---
name: ci-proyecto
description: Cómo redactar o corregir la integración continua del proyecto (su workflow de GitHub Actions) a partir del perfil, para que ejecute los mismos verbos que el equipo. Usar con el comando bowser-ci, cuando el proyecto suma una parte, un verbo o un servicio, o cuando cambia la versión de una herramienta.
---
# La integración continua del proyecto

La integración continua (CI) ejecuta en GitHub, en cada Pull Request, lo que el equipo ejecuta en su máquina con `make ci`. Es la única revisión que nadie puede saltarse.

Hay dos workflows, y no se mezclan:

| Archivo | De quién es | Qué trae |
|---|---|---|
| `.github/workflows/bowser.yml` | Del kit. **No se edita** | Lo que no depende de la tecnología: archivos del kit, agentes y perfil al día, archivos inmutables, secretos |
| `.github/workflows/ci.yml` | Del proyecto. Lo redactas tú | Lo que sí depende: instalar las herramientas de cada parte, levantar sus servicios y ejecutar sus verbos |

## La regla central

**El workflow no repite los comandos de los verbos: los ejecuta.** Cada parte se revisa con una sola línea:

```yaml
- run: make ci PART=<nombre de la parte>
```

Así el Pull Request se revisa con los mismos comandos que usa el equipo, y cambiar un comando en el perfil cambia la CI sin tocar el workflow. Todo lo demás del archivo existe para que esa línea pueda correr: el lenguaje instalado, las herramientas que los verbos llaman y los servicios que las pruebas necesitan.

## Cómo redactarlo

1. **Parte de los hechos.** `make profile` (partes, carpetas y el comando de cada verbo) y `make check-ci` (qué ejecuta ya algún workflow). Lee los comandos de los verbos: cada programa que nombran tiene que estar instalado en el job. Los `requisitos` del perfil (al final de `make profile`) son la lista de lo que el equipo instala en su máquina, con sus versiones mínimas: sirve de guía, pero el job instala solo lo que los verbos de su parte usan.
2. **Un job por parte que tenga al menos un verbo.** Una parte sin verbos no lleva job: lístala en la entrega como "sin revisar". Dos partes que usan lo mismo pueden compartir job si eso lo hace más corto; sepáralas si una necesita un servicio que la otra no.
3. **En cada job, en este orden:**
   - Descargar el código.
   - Instalar el lenguaje **en la versión del proyecto**, leída de su propio archivo de versión (el que use su tecnología) y no escrita a mano en el workflow: así no se separan.
   - Instalar las dependencias de la parte con el comando reproducible de su gestor (el que respeta el archivo de bloqueo).
   - Instalar las herramientas que los verbos llaman y el lenguaje no trae (linters, generadores, auditores), cada una en una versión exacta.
   - Levantar los servicios que las pruebas necesitan (base de datos, cola) como `services:` del job, con la misma versión mayor que el entorno local, una comprobación de salud y credenciales de usar y tirar.
   - Preparar el servicio si hace falta (por ejemplo, aplicar los archivos de cambio a la base de datos vacía). Eso además comprueba que se aplican desde cero.
   - `make ci PART=<nombre>`.
4. **Cuándo corre:** en cada Pull Request y en cada cambio que llega a la rama principal. Con `concurrency` para cancelar la ejecución anterior de la misma rama.
5. **Permisos mínimos:** `permissions: contents: read`, salvo que un paso necesite más y digas cuál y por qué.
6. **Versiones fijas, nunca "la última".** Las acciones, en su versión mayor (`@v7`); toma como referencia las que usa `bowser.yml`, que el kit mantiene al día. Las herramientas, en versión exacta: la que el proyecto ya declara en sus archivos o la que el equipo tiene instalada. Si no puedes saber una versión, **pregúntala**; no la inventes ni la supongas de memoria.
7. **Secretos.** Nada real dentro del archivo: lo que haga falta va en los secretos del repositorio, y lo dices en la entrega. La contraseña de una base de datos que nace y muere con el job no es un secreto.
8. **Si el proyecto ya tiene `ci.yml`, se actualiza.** Conserva sus pasos propios. Si repite lo que ya hace `bowser.yml` (verificar el kit o los agentes, buscar secretos), propón quitarlo: correría dos veces.

## Lo que no va aquí

- **Comandos de verbos copiados.** Si una comprobación debe correr en la CI, primero es un verbo del perfil (o parte de uno); si no encaja en ninguno, es un paso propio del workflow y lo explicas.
- **Despliegues.** Van en otro workflow, y siempre con aprobación de una persona.
- **Cambios en `bowser.yml`**, en el perfil o en los comandos de los verbos. Si un verbo no puede correr en GitHub tal como está (por ejemplo, depende de algo que solo existe en una máquina), dilo: ese cambio es del perfil y lo confirma una persona.

## Cómo se comprueba

- `make check-ci` debe terminar diciendo que la CI ejecuta todos los verbos que el perfil define.
- `make ci` debe pasar en local antes: lo que falla en la máquina falla en GitHub.
- Un workflow **solo se prueba de verdad en GitHub**. Dilo en la entrega, y revisa la primera ejecución con la persona.

## Entrega

- El archivo propuesto.
- Un resumen en lenguaje simple: qué se revisa en cada parte, qué se instala y en qué versión, qué servicios se levantan.
- Las partes y los verbos que quedan sin revisar, y por qué.
- Los secretos o variables que una persona debe crear en GitHub.
- Las versiones que no pudiste confirmar.
