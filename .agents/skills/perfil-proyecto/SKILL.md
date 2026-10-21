---
name: perfil-proyecto
description: Cómo redactar o corregir el perfil del proyecto (equipo/perfil.json), que declara sus partes, tecnologías y el comando de cada verbo. Usar con el comando bowser-profile, al iniciar un proyecto, o cuando el proyecto suma una parte o una tecnología.
---
# El perfil del proyecto

`equipo/perfil.json` describe el proyecto para que el kit no suponga nada: qué partes tiene, dónde están y cómo se ejecuta cada verbo. Es del proyecto. Lo redacta un agente y lo aprueba una persona.

## Formato

```json
{
  "partes": [
    {
      "nombre": "api",
      "carpeta": "backend",
      "descripcion": "API REST del producto",
      "roles": [
        {"rol": "dev-backend", "skills": ["go-backend"]},
        {"rol": "dev-dba", "skills": ["postgres-db"]}
      ],
      "terceros": ["vendor"],
      "inmutables": ["migrations/*.sql"],
      "verbos": {
        "formato": "test -z \"$(gofmt -l .)\"",
        "revisar": "go vet ./... && golangci-lint run",
        "probar": "go test ./...",
        "cobertura": null,
        "auditar": "govulncheck ./...",
        "generar": null
      }
    }
  ],
  "requisitos": [
    {"comando": "go", "nombre": "Go", "minima": "1.26", "version": "go version"},
    {"comando": "golangci-lint", "opcional": true, "instalar": "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0"}
  ]
}
```

| Campo | Qué es |
|---|---|
| `nombre` | Minúsculas, números y guiones. Único |
| `carpeta` | Relativa a la raíz. `"."` si el proyecto es una sola parte |
| `roles` | Los agentes de `equipo/agentes/` que trabajan esa parte. Cada elemento es un nombre (`"dev-backend"`) o un objeto con las skills que son solo de ese rol (`{"rol": "dev-frontend", "skills": ["react-frontend"]}`) |
| `skills` | Skills con las convenciones de su tecnología: del catálogo del kit o propias del proyecto (`.agents/skills/`). Las reciben todos los roles de la parte, además de las propias de cada uno |
| `terceros` | Carpetas con código ajeno copiado dentro de la parte (relativas a ella). No se revisan ni se formatean |
| `inmutables` | Patrones de archivos que no se modifican una vez versionados (relativos a la parte). Ej.: migraciones |
| `verbos` | El comando de cada verbo. Se ejecuta con `bash` **dentro de la carpeta de la parte** |

## Los verbos

| Verbo | Qué debe hacer el comando | Lo usa |
|---|---|---|
| `formato` | Comprobar el formato sin cambiar archivos. Rápido | Cada commit (si el commit toca la parte) y `make lint` |
| `revisar` | Análisis estático: linters, tipos, sintaxis | `make lint` |
| `probar` | Ejecutar las pruebas | `make test` |
| `cobertura` | Medir cobertura y **fallar** bajo el mínimo del proyecto | `make coverage` |
| `auditar` | Buscar vulnerabilidades en las dependencias | `make security` |
| `generar` | Regenerar el código generado | `make generate`, `make check-generated` |

Todo comando termina con código 0 si está bien y distinto de 0 si no.

Cada comando recibe dos variables de entorno: `PERFIL_RAIZ` (ruta completa de la raíz del proyecto) y `PERFIL_TERCEROS` (las carpetas de `terceros` de esa parte, una por línea). Úsalas en vez de repetir la lista de terceros dentro del comando.

Si un verbo necesita un script propio del proyecto, va en `tools/`, en la raíz, y el comando lo llama con `bash "$PERFIL_RAIZ/tools/<script>.sh"`. Nunca en `scripts/`: es del kit. `tools/` no es una parte del proyecto, aunque la detección la liste como carpeta con código.

## Los requisitos

`requisitos` va fuera de `partes`: es del proyecto entero. Lista los programas que una máquina necesita para trabajar en él, y es lo que revisa `make doctor`. Sin esa lista, `make doctor` no revisa nada de la tecnología del proyecto.

| Campo | Qué es |
|---|---|
| `comando` | El programa, tal como se escribe en la terminal (`go`, `php`, `docker`). Sin espacios ni rutas. El único obligatorio |
| `nombre` | Cómo se muestra. Si falta, el del comando |
| `minima` | Versión mínima, entre comillas (`"1.26"`) |
| `version` | El comando que imprime la versión, si no es `<comando> --version` (ej. `go version`, `java -version 2>&1`) |
| `opcional` | `true`: si falta es un aviso y no un problema. Para lo que solo algunos del equipo usan |
| `instalar` | Cómo se instala, en una línea. Se muestra cuando falta |

Un elemento puede ser solo el nombre del programa (`"php"`).

Cómo llenarla:

- **Sale de los verbos y de cómo se levanta el proyecto.** Cada programa que un comando de `verbos` llama es un requisito, y también lo que haga falta para levantar el entorno local. Si los verbos corren dentro de Docker, el requisito es `docker`, no el lenguaje.
- **No repitas lo que el kit ya revisa siempre:** `git`, `python3`, `make`, `jq`, `uv`, `specify`, `gitleaks` y el agente de código. Docker tampoco hace falta si el proyecto tiene archivo de Docker Compose: `make doctor` lo revisa solo.
- **La versión mínima sale del proyecto**, no de tu memoria: de su archivo de versión (`go.mod`, `engines` de `package.json`, `composer.json`, `.tool-versions`). Si no está escrita en ningún lado, pregúntala; si la persona no la sabe, deja el requisito sin `minima`.
- **Ejecuta el comando de versión antes de darlo por bueno.** `make doctor` toma el primer número con forma de versión que aparece después del nombre del programa (así lee bien `govulncheck -version`, que imprime primero la de Go); si ninguna línea lo nombra, el primero que imprima. Hay programas que no aceptan `--version` (`go version`, `govulncheck -version`). Si la versión que sale no es la del programa, cambia el comando de `version`; si ninguno la da, deja el requisito sin `minima`: se comprueba solo que esté instalado.
- `instalar` solo con un comando que conozcas con certeza o que el proyecto ya documente. Ninguno es mejor que uno inventado.
- Obligatorio es lo que hace falta para que `make ci` corra entero; lo demás, `opcional`.

## Cómo redactarlo

1. Parte de los hechos: la salida de `make profile DETECT=1` y lo que leas en el código (archivos de configuración, `README`, scripts existentes). Un archivo de tecnología que trae `dentro_de_carpeta_a_confirmar` suele ser de una biblioteca copiada dentro del proyecto: no lo tomes por la tecnología del proyecto sin mirar la carpeta.
2. **Una parte es algo que se construye, se prueba o se despliega por separado.** No dividas por dividir: un proyecto pequeño puede ser una sola parte con `"carpeta": "."`. Si una parte mezcla oficios (por ejemplo, PHP que genera HTML junto a su CSS y su JavaScript), no la partas en carpetas artificiales: dale varios `roles`, cada uno con sus skills.
   **Quién trabaja cada parte.** Los roles no traen tecnología: reciben del perfil sus carpetas, skills y comandos, y una parte sin `roles` no tiene quién la trabaje. `dev-backend`, el lado del servidor; `dev-frontend`, la interfaz; `dev-dba`, la base de datos (esquema, migraciones, consultas, procedimientos). **La base de datos es de `dev-dba`, no de `dev-backend`:** si vive en su propia carpeta, es una parte con ese rol; si sus archivos están dentro de la carpeta del servidor, esa parte lleva los dos roles y la skill del motor va en `dev-dba`. Cualquier otro rol de `roles_disponibles` puede trabajar una parte si el proyecto lo necesita (por ejemplo, `devops` en una carpeta de infraestructura).
3. Usa solo comandos que el proyecto ya puede ejecutar: herramientas que ya usa o que están en sus archivos de configuración. Si propones una herramienta nueva, dilo aparte como recomendación; no la pongas en el perfil hasta que esté instalada.
4. **Si el proyecto no tiene algo, el verbo va en `null`.** Un proyecto sin pruebas tiene `"probar": null`, no un comando inventado que siempre pasa. Lista los verbos sin definir como deuda, para que la persona decida.
5. Lo que no puedas deducir, pregúntalo: cómo se levanta, qué versión del lenguaje usa, si una carpeta dudosa es de terceros.
6. Asigna solo skills que aparezcan en `skills_disponibles` de la detección: son las del catálogo del kit más las propias del proyecto. Las del catálogo que el perfil nombre las copia `make install-kit`, y las que deje de nombrar las retira. Si una es `redactada` (ver `madurez_de_las_skills_del_catalogo`), avisa que todavía no se usó en un proyecto real. Si falta la de una tecnología, dilo: el agente trabajará sin ella.
7. Las skills del catálogo dicen el estándar de cada tecnología, no cómo lo hace este proyecto. Lo propio del proyecto (nombres, estructura, patrones, bibliotecas) va en una skill `convenciones-de-<parte>`, que redacta el comando `bowser-conventions` y se nombra aquí junto a la del catálogo. No la inventes al redactar el perfil: si falta, `make profile` lo recuerda.
8. Nada de secretos en los comandos: las credenciales vienen de variables de entorno.

## Lo que el kit no trae y el perfil permite completar

El kit no instala nada que dependa de la tecnología. Al redactar el perfil, revisa y propone junto con él:

- **`.gitignore`.** El bloque del kit solo cubre secretos, editores y sus propios archivos. Lo que cada tecnología genera lo dice la sección "Lo que no va en git" de su skill: escribe esas reglas en el `.gitignore` del proyecto, **fuera del bloque del kit** y con la carpeta de cada parte delante (`web/node_modules/`). Antes mira `git status`: si algo de eso ya está versionado, no lo saques de git por tu cuenta; avísalo.
- **Entorno local.** `make up` levanta lo que diga el archivo de Docker Compose de la raíz (`compose.yaml` o `docker-compose.yml`). El kit no trae ninguno. Si el proyecto no lo tiene, dilo en la entrega como pendiente: crearlo es tarea de `devops`, con los servicios y las versiones que el perfil y las skills nombran. No todos los proyectos lo necesitan.
- **`.env.example`.** Lista las variables de entorno que el proyecto necesita, sin valores reales. El kit lo deja vacío: se llena cuando existe la primera variable.

## Entrega

- El JSON propuesto.
- Un resumen en lenguaje simple: partes, tecnología de cada una, roles que la trabajan, verbos definidos y verbos sin definir, y lo que cada máquina debe tener instalado.
- Las preguntas abiertas.

Después de escribirlo, `make profile` lo valida, `make doctor` dice si la máquina cumple los requisitos y `make install-kit` trae sus skills y regenera los roles con lo que el perfil dice de cada parte (sin ese paso, el commit se rechaza porque los archivos generados no están al día). El commit lo confirma una persona con `APPROVED_PROFILE=1`; un agente nunca usa esa variable.
