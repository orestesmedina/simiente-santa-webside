# Guía de inicio

Guía para personas nuevas en el equipo. Explica cómo instalar el entorno y cómo trabajar una funcionalidad de principio a fin con el equipo de agentes.

**Tiempo estimado:** instalación, 1 a 2 horas. Primera funcionalidad de práctica, medio día.

---

## Contenido

1. [Cómo funciona esto (léelo primero)](#1-cómo-funciona-esto-léelo-primero)
2. [Accesos que necesitas](#2-accesos-que-necesitas)
3. [Instalación del entorno](#3-instalación-del-entorno)
4. [Preparar un proyecto](#4-preparar-un-proyecto)
5. [Flujo de desarrollo de principio a fin](#5-flujo-de-desarrollo-de-principio-a-fin)
6. [Otros flujos: bugs y cambios pequeños](#6-otros-flujos-bugs-y-cambios-pequeños)
7. [Reglas de oro](#7-reglas-de-oro)
8. [Problemas comunes](#8-problemas-comunes)
9. [Glosario](#9-glosario)
10. [Lista de la primera semana](#10-lista-de-la-primera-semana)

---

## 1. Cómo funciona esto (léelo primero)

En esta empresa **la IA escribe casi todo el código, pero una persona es responsable del resultado**. Esa persona eres tú.

Tu trabajo no es escribir código. Es:

- **Entender** lo que el cliente necesita y asegurarte de que la especificación lo refleje.
- **Aprobar o rechazar** la especificación y el plan técnico antes de que se construya nada.
- **Revisar** el Pull Request final y decidir si se integra.
- **Intervenir** cuando los agentes se atascan o se desvían.

### El equipo de agentes

Trabajas con un agente principal (el **orquestador**) que coordina a 10 especialistas. **El orquestador es el agente con el que hablas** al abrir la herramienta en la carpeta del proyecto (en OpenCode aparece con el nombre `orquestador`). Le hablas con naturalidad ("construyamos la funcionalidad 2 del roadmap", "hay un error al guardar"), y él decide qué flujo seguir, usa Spec Kit y delega en los especialistas:

| Rol | Qué hace | Puede escribir código |
|---|---|---|
| `analista-producto` | Convierte la necesidad en `spec.md` | No (solo documentos) |
| `arquitecto` | Diseña `plan.md`, modelo de datos y API; divide en tareas | No (solo documentos) |
| `disenador-ux` | Define pantallas, flujos y estados en `ux.md` | No (solo documentos) |
| `dev-backend` | Implementa en Go y PostgreSQL | Sí |
| `dev-frontend` | Implementa en React + TypeScript | Sí |
| `qa-tester` | Prueba contra los criterios de aceptación | Solo pruebas |
| `revisor-codigo` | Revisa calidad y apego al plan | No |
| `seguridad` | Audita vulnerabilidades y secretos | No |
| `devops` | Docker, CI/CD, despliegues | Sí (infraestructura) |
| `documentador` | README, CHANGELOG, notas para el cliente | No (solo documentos) |

**Principio clave: quien escribe el código nunca lo aprueba.**

### El proceso: Spec-Driven Development

Todo empieza por escrito, antes del código:

```
Idea del cliente
   │
   ▼
spec.md ─────── ✋ TÚ APRUEBAS  (¿es lo que el cliente pidió?)
   │
   ▼
plan.md ─────── ✋ TÚ APRUEBAS  (¿la solución técnica tiene sentido?)
   │
   ▼
tasks.md → implementación → QA + revisión + seguridad → corrección (bucle)
   │
   ▼
Pull Request ── ✋ TÚ APRUEBAS  (¿está bien hecho? ¿se integra?)
   │
   ▼
Producción ──── ✋ TÚ APRUEBAS
```

Si la spec y el plan están bien, la implementación casi siempre sale bien. **La mayor parte de tu atención va en esas dos aprobaciones.**

### Archivos que debes conocer

| Archivo | Qué es | ¿Lo editas? |
|---|---|---|
| `AGENTS.md` | Instrucciones del orquestador | Solo con aprobación de dirección técnica |
| `.specify/memory/constitution.md` | Reglas no negociables del código | Solo con aprobación de dirección técnica |
| `equipo/agentes/*.md` | Definición de cada rol | Solo con aprobación de dirección técnica |
| `.agents/skills/` | Convenciones del stack y flujos | Solo con aprobación de dirección técnica |
| `specs/<número>-<feature>/` | Spec, plan y tareas de cada funcionalidad | Sí, es tu día a día |
| `specs/<número>-<feature>/estado.md` | Fase, aprobaciones, hallazgos, decisiones y próximo paso de la funcionalidad | Lo mantiene el orquestador; puedes corregirlo |
| `specs/<número>-<feature>/costos.json` | Tokens y costo de IA por agente y modelo | **Nunca** a mano: lo escribe `make costos` |
| `docs/producto/roadmap.md` | Funcionalidades del producto y su estado | Tú decides orden y alcance; el orquestador actualiza el estado |
| `CLAUDE.md`, `.claude/`, `.codex/`, `.opencode/`, `opencode.json` | Archivos generados | **Nunca** a mano |
| `.bowser-spec-kit-ai/` y `.kit-manifest.json` | El kit compartido (submódulo) y la lista de archivos que vienen de él | **Nunca** a mano; se actualizan con `make actualizar-kit` |
| `proyecto.mk` | Comandos de `make` propios del proyecto | Sí |
| `equipo/config.json` | Qué modelo usa cada agente (ver `equipo/MODELOS.md`) | Solo con aprobación de dirección técnica |

---

## 2. Accesos que necesitas

Pide esto a tu líder antes de empezar:

- [ ] Cuenta de **GitHub** con acceso a la organización y a los repositorios de los proyectos.
- [ ] Acceso al **agente de código** que usa el equipo: una suscripción o clave de API de Claude Code, Codex u OpenCode. Pregunta cuál está activo en `equipo/config.json`.
- [ ] Valores del archivo `.env` de cada proyecto (se entregan por un gestor de contraseñas, **nunca** por chat ni correo).
- [ ] Acceso al gestor de tareas del equipo y al canal de comunicación.

---

## 3. Instalación del entorno

### 3.1 Sistema operativo

- **macOS** o **Linux**: funcionan directamente.
- **Windows**: se trabaja dentro de **WSL2 con Ubuntu**. Lee la sección 3.1.1 completa antes de instalar nada: la mayoría de los problemas en Windows vienen de mezclar los dos "mundos".

#### 3.1.1 Windows: cómo funciona WSL (importante)

En Windows con WSL conviven **dos sistemas separados, cada uno con sus propios programas**:

| | Windows | Ubuntu (WSL) |
|---|---|---|
| Qué va aquí | Navegador, DBeaver, Docker Desktop, VS Code (la ventana) | `make`, `python3`, `jq`, Go, Node, Spec Kit, git, **el agente de código** y **los proyectos** |
| Terminal | PowerShell | Terminal "Ubuntu" |

Un programa instalado en uno **no existe** en el otro. Por eso la regla es: **todo lo del desarrollo se instala y se ejecuta dentro de Ubuntu**.

**Pasos:**

1. Instala WSL con Ubuntu (PowerShell como administrador) y reinicia:
   ```powershell
   wsl --install -d Ubuntu
   ```
2. Abre la terminal **Ubuntu**. Desde aquí, **todos** los comandos de esta guía se ejecutan en esa terminal, salvo que se diga lo contrario.
3. **Guarda los proyectos dentro de Ubuntu**, por ejemplo en `~/proyectos/`. No trabajes en `/mnt/c/...` (el disco de Windows visto desde Ubuntu): es mucho más lento y da problemas con los permisos de los scripts.
4. **Docker:** instala [Docker Desktop](https://www.docker.com/products/docker-desktop/) en Windows y actívalo para Ubuntu en **Settings → Resources → WSL Integration → Ubuntu → Apply & Restart**. Docker Desktop debe estar abierto cuando trabajes. Verifica en Ubuntu con `docker version`.
5. **VS Code:** instala la extensión **WSL** de Microsoft y abre siempre los proyectos desde la terminal de Ubuntu:
   ```bash
   cd ~/proyectos/mi-proyecto
   code .
   ```
   Abajo a la izquierda debe decir **"WSL: Ubuntu"**. Si abres la carpeta desde Windows, VS Code usa el Git de Windows y los commits fallan con mensajes confusos, porque ahí no hay `python3` ni `jq`.
6. **El agente de código** (Claude Code, Codex u OpenCode) se instala y se ejecuta **en Ubuntu** (sección 3.3). El agente ejecuta comandos como `make test` o `git commit`, y los ejecuta en el sistema donde está instalado. Si ya lo tenías instalado en Windows, lee la sección 3.1.2.
7. **PostgreSQL:** **no** instales PostgreSQL en Windows; el proyecto lo levanta con Docker. Si ya tienes uno instalado, ocupa el puerto 5432 y tu cliente se conectará a ese en lugar del de Docker (ver problemas comunes).
8. **Cliente de base de datos:** usa uno actualizado que soporte PostgreSQL 16, como [DBeaver Community](https://dbeaver.io/download/) (gratuito) o pgAdmin 4. Los clientes viejos (por ejemplo Navicat 11) no pueden autenticarse con PostgreSQL moderno.

#### 3.1.2 Windows: cuando Ubuntu usa por error un programa de Windows

WSL agrega por defecto las rutas de Windows al `PATH` de Ubuntu. Por eso, si tenías un programa instalado en Windows (por ejemplo OpenCode, Node o Git instalados con npm o con un instalador de Windows), **Ubuntu lo encuentra y lo usa aunque no esté instalado en Ubuntu**. Parece que funciona, pero cuando ese programa ejecuta comandos, lo hace del lado de Windows, donde no están `make`, `python3` ni el resto de herramientas.

`make doctor` lo detecta y lo marca con ✗. Para revisarlo a mano:

```bash
which -a opencode      # o claude, codex, node, git...
```

- Rutas que empiezan con `/home/...` o `/usr/...` son de **Ubuntu** (bien).
- Rutas que empiezan con `/mnt/c/...` son de **Windows** (mal, si es la primera de la lista).

**Cómo corregirlo (ejemplo con OpenCode):**

1. Instálalo dentro de Ubuntu:
   ```bash
   curl -fsSL https://opencode.ai/install | bash
   ```
2. Cierra y vuelve a abrir la terminal, y verifica que la **primera** línea de `which -a opencode` empiece con `/home/`. Si sigue apareciendo primero la de Windows, pon la carpeta que indicó el instalador (normalmente `~/.opencode/bin`) al inicio del `PATH`:
   ```bash
   echo 'export PATH="$HOME/.opencode/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   ```
3. **Vuelve a iniciar sesión / configurar el proveedor.** La instalación de Ubuntu no comparte configuración con la de Windows: abre el agente en tu proyecto y conecta tu cuenta de nuevo (en OpenCode, con `/connect`; en Claude Code y Codex, al iniciar te pide entrar).
4. **Opcional:** si no usas ese programa desde Windows, desinstálalo de Windows para evitar confusiones (por ejemplo, en PowerShell: `npm uninstall -g opencode-ai`).

No desactives la integración de rutas de Windows en WSL: de ella depende, entre otras cosas, que `code .` abra VS Code desde Ubuntu.

### 3.2 Herramientas base

**macOS** (con [Homebrew](https://brew.sh)):
```bash
brew install git go node@24 python@3.12 jq make gh gitleaks golang-migrate
brew install --cask docker        # Docker Desktop; ábrelo una vez para que arranque
```

**Ubuntu / WSL2** (en la terminal de Ubuntu):
```bash
sudo apt update && sudo apt install -y git jq make unzip curl build-essential python3 python3-pip
# Go 1.26+ : https://go.dev/doc/install
# Node 24  : https://github.com/nvm-sh/nvm  →  nvm install 24
# Docker   : en Windows, Docker Desktop con integración WSL (ver 3.1.1)
#            en Linux nativo: https://docs.docker.com/engine/install/ubuntu/
# GitHub CLI: https://github.com/cli/cli/blob/trunk/docs/install_linux.md
```

`jq` lee archivos JSON y lo usan los hooks que protegen archivos; `make` ejecuta los comandos del proyecto (`make up`, `make doctor`…). Sin ellos, los controles no funcionan.

**Antes de seguir, ejecuta el diagnóstico** desde la carpeta de un proyecto que ya tenga el kit instalado: `make doctor` (o `bash scripts/doctor.sh` si todavía no tienes `make`). Te muestra de una vez todo lo que falta.

**uv y Spec Kit** (todos los sistemas):
```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
uv tool install specify-cli
specify --version
```

**Herramientas de Go** (todos los sistemas; las mismas versiones que fija el CI en `.github/workflows/ci.yml`):
```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0
go install golang.org/x/vuln/cmd/govulncheck@v1.8.0
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.20.1
```
Asegúrate de que `$(go env GOPATH)/bin` esté en tu `PATH`.

### 3.3 El agente de código

Instala **el que use el equipo** (revisa `"herramientas"` en `equipo/config.json`). En Windows, **instálalo en la terminal de Ubuntu**, no en Windows (ver 3.1.1):

| Herramienta | Instalación | Primer inicio |
|---|---|---|
| Claude Code | `curl -fsSL https://claude.ai/install.sh \| bash` | `claude` → inicia sesión |
| Codex | `npm install -g @openai/codex` | `codex` → inicia sesión |
| OpenCode | `curl -fsSL https://opencode.ai/install \| bash` | `opencode` → configura el proveedor |

Si un comando de instalación falla, consulta la documentación oficial de la herramienta; estos métodos pueden cambiar.

### 3.4 GitHub

```bash
gh auth login              # los agentes lo usan para abrir Pull Requests
git config --global user.name  "Tu Nombre"
git config --global user.email "tu@empresa.com"
```

---

## 4. Preparar un proyecto

### Caso A: unirte a un proyecto existente (lo más común)

```bash
git clone --recursive git@github.com:<organizacion>/<proyecto>.git   # --recursive trae el kit (.bowser-spec-kit-ai/)
cd <proyecto>

make instalar-hooks          # activa los controles de git (obligatorio)
cp .env.example .env         # y completa los valores que te entregaron
make up                      # levanta PostgreSQL
make doctor                  # verifica que todo esté listo
```

> **Importante:** ajusta la contraseña en `.env` **antes** del primer `make up`. PostgreSQL solo la toma al crear la base; si la cambias después, hay que recrearla con `docker compose down -v && make up` (borra los datos locales).

#### Conectarte a la base de datos local

Con un cliente como DBeaver o pgAdmin 4 (ver 3.1.1 si estás en Windows):

| Campo | Valor |
|---|---|
| Host | `localhost` |
| Puerto | `POSTGRES_PORT` de tu `.env` (por defecto `5432`) |
| Base de datos | `POSTGRES_DB` (por defecto `app`) |
| Usuario / contraseña | `POSTGRES_USER` / `POSTGRES_PASSWORD` de tu `.env` |

Para comprobarlo desde la terminal: `docker compose exec db psql -U app -d app -c "select version();"`

`make doctor` debe terminar con **"Todo listo para trabajar"**. Si marca problemas, revisa la [sección 8](#8-problemas-comunes).

Luego lee, en este orden:
1. `README.md` del proyecto.
2. `.specify/memory/constitution.md`.
3. `AGENTS.md`.
4. La carpeta `specs/`, empezando por la funcionalidad más reciente, para ver cómo se ha trabajado.

**Sobre la carpeta `.bowser-spec-kit-ai/`:** es el kit compartido de la empresa, incluido como submódulo de git. Sus archivos se copian a la raíz del proyecto con `make instalar-kit`. **No edites `.bowser-spec-kit-ai/` ni los archivos que vienen de él** (están listados en `.kit-manifest.json`): el hook de git y el CI lo detectan. Si algo del kit debería cambiar, propónlo a dirección técnica.

### Caso B: crear un proyecto nuevo

Solo lo hace quien tenga autorización. Sigue la sección **Instalación** del `README.md` del kit: instalar Spec Kit, `specify init`, `git submodule add … .bowser-spec-kit-ai` y `make -f .bowser-spec-kit-ai/Makefile instalar-kit`.

### De la idea a la primera funcionalidad (proyecto nuevo)

No le pidas al agente "hazme la aplicación". Primero se convierte la idea en una lista de funcionalidades pequeñas, y luego se construyen una por una con el flujo de la sección 5.

1. **Escribe la idea en una página, tú mismo:**
   ```bash
   mkdir -p docs/producto
   cp docs/plantillas/idea.md docs/producto/idea.md
   ```
   Complétala sin IA: problema, usuarios, qué sería un éxito, lo mínimo que debe hacer, qué queda fuera y restricciones. Si no puedes llenar una sección, es una pregunta para el cliente.
2. **Pide el MVP dividido en funcionalidades:**
   > Lee docs/producto/idea.md. Propón el MVP más pequeño y divídelo en funcionalidades independientes, en orden de dependencia. La primera debe ser la estructura base del proyecto. Guárdalo en docs/producto/roadmap.md con el formato de docs/plantillas/roadmap.md. No escribas specs todavía.

   Revisa y ajusta la lista: ese orden es el plan del proyecto.
3. **Primera funcionalidad, el esqueleto:**
   > Usa la skill equipo-feature: estructura base del proyecto. Backend en Go con endpoint /healthz conectado a PostgreSQL, frontend React que muestre el estado del backend, Docker Compose levantando todo, y CI en verde.

   Los agentes copian los patrones del código existente, así que un esqueleto limpio y aprobado hace que todo lo siguiente salga consistente.
4. **Sigue con el roadmap**, una funcionalidad a la vez, integrando cada una antes de empezar la siguiente. La columna **Estado** del roadmap muestra en todo momento qué está terminado, en curso o pendiente.

### Actualizar el kit en un proyecto

Cuando dirección técnica publique una mejora del kit:

```bash
make actualizar-kit
make actualizar-modelos   # solo si el paso anterior avisó "el kit recomienda modelos distintos"
git add . && git commit -m "chore: actualiza kit de desarrollo"
```

Al terminar, `make actualizar-kit` muestra **qué cambió** entre tu versión y la nueva, y al final los **pasos manuales** si los hay: hazlos antes del commit. Para ver el historial completo en cualquier momento: `make novedades` (o `make novedades DESDE=1.4.0`).

Los modelos de cada agente viven en `equipo/config.json`, que es del proyecto: actualizar el kit no los cambia. `make actualizar-modelos` te muestra qué cambiaría, pide confirmación y regenera la configuración de agentes (incluido `.opencode/`). Después reinicia OpenCode (o tu herramienta) para que tome los modelos nuevos.

Si el comando se detiene porque un archivo del kit fue modificado en el proyecto, no uses `FORZAR=1` sin consultar: ese cambio local podría ser importante.

---

## 5. Flujo de desarrollo de principio a fin

Usaremos un ejemplo: **"Los clientes pueden registrarse con email y contraseña."**

### Paso 0 — Antes de empezar

```bash
git checkout main && git pull
make up
make doctor
make estado      # qué está en curso y qué sigue en el roadmap
```
Abre tu agente en la carpeta del proyecto (`claude`, `codex` u `opencode`). Lo primero que hace el orquestador es decirte dónde quedó el trabajo.

### Paso 1 — Pedir la funcionalidad

Escríbele al agente:

> Usa la skill equipo-feature para esta funcionalidad: los clientes pueden registrarse con email y contraseña. Deben confirmar su email antes de poder iniciar sesión. Si el email ya existe, deben ver un mensaje claro.

**Consejos para una buena solicitud:**
- Describe **qué** y **para quién**, no **cómo**. Nada de "usa tal librería".
- Incluye reglas de negocio que conozcas: límites, casos especiales, qué pasa si algo falla.
- Adjunta lo que tengas del cliente: correos, notas de reunión, capturas.
- Si no sabes algo, dilo. El analista te hará preguntas.

Spec Kit crea una carpeta para la funcionalidad (por ejemplo `specs/001-registro-usuarios/`) y normalmente una rama de git con el mismo nombre.

### Paso 2 — Responder aclaraciones

El `analista-producto` puede marcar dudas como `[NECESITA ACLARACIÓN]` y preguntarte. Respóndelas tú o consúltalas con el cliente. **No dejes que el agente adivine reglas de negocio.**

### Paso 3 — ✋ Aprobar la especificación

Abre `specs/001-registro-usuarios/spec.md` y revísalo con esta lista:

- [ ] Describe lo que el cliente pidió, ni más ni menos.
- [ ] Cada historia de usuario tiene criterios de aceptación en formato Dado / Cuando / Entonces.
- [ ] Los criterios son verificables. "Rápido" no sirve; "responde en menos de 2 segundos" sí.
- [ ] Incluye casos de error y casos límite (email duplicado, contraseña débil, enlace vencido…).
- [ ] La sección "fuera de alcance" es correcta.
- [ ] No quedan marcas `[NECESITA ACLARACIÓN]`.
- [ ] No menciona tecnología (eso va en el plan).

Si algo falla, pide cambios concretos: *"Agrega un criterio: la contraseña debe tener mínimo 12 caracteres."*
Cuando esté bien, escribe: **"Apruebo la spec."**

### Paso 4 — ✋ Aprobar el plan técnico

El `arquitecto` genera `plan.md`, `data-model.md` y `contracts/`. Si hay pantallas, el `disenador-ux` genera `ux.md`. Revisa:

- [ ] Respeta el stack oficial (React, Go, PostgreSQL) y la constitución.
- [ ] Cada requisito de la spec tiene una parte del plan que lo resuelve.
- [ ] El modelo de datos tiene sentido: tablas, relaciones, restricciones e índices.
- [ ] Los endpoints de la API cubren todo lo que el frontend necesita.
- [ ] Las dependencias nuevas están justificadas.
- [ ] Los riesgos están identificados.
- [ ] (Si hay UI) `ux.md` cubre estados de carga, vacío, error y éxito.

Si no tienes experiencia técnica suficiente para juzgar el plan, **pide revisión a alguien senior antes de aprobar**. No hay vergüenza en eso; aprobar un mal plan sí es caro.

Cuando esté bien: **"Apruebo el plan."**

### Paso 5 — Tareas y verificación de coherencia (automático)

El `arquitecto` genera `tasks.md` y el `revisor-codigo` verifica que spec, plan y tareas sean coherentes. Dale un vistazo a `tasks.md`: las tareas deben ser pequeñas y cada una debe incluir sus pruebas.

### Paso 6 — Implementación (automático, con supervisión)

Los desarrolladores implementan tarea por tarea, con un commit por tarea. Tu papel aquí:

- **Deja trabajar**, pero revisa el progreso cada cierto tiempo.
- **Intervén si ves** que el agente repite el mismo error, modifica cosas fuera del alcance o dice "voy a desactivar esta prueba".
- **Si un commit es rechazado** por los hooks de git, el agente debe corregir la causa. Nunca le permitas usar `--no-verify`.

### Paso 7 — Validación (automático)

`qa-tester`, `revisor-codigo` y `seguridad` revisan en paralelo. Si alguno rechaza, el trabajo vuelve al desarrollador. Después de 3 ciclos sin éxito, el orquestador se detiene y te pide ayuda. En ese caso:

1. Lee los reportes en `specs/<feature>/revision-*.md`.
2. Decide si el problema es de código (da instrucciones más precisas), de plan (vuelve al paso 4) o de spec (vuelve al paso 3).

### Paso 8 — ✋ Revisar y aprobar el Pull Request

El orquestador abre un PR. Antes de aprobar:

- [ ] El CI de GitHub está en verde.
- [ ] Los reportes de QA, revisión y seguridad dicen APROBADO.
- [ ] Probaste la funcionalidad tú mismo en local (`make up`, levantar backend y frontend, recorrer el flujo).
- [ ] Los cambios tienen sentido al leer el diff; no hay archivos inesperados.
- [ ] El CHANGELOG y la documentación están actualizados.
- [ ] No hay secretos, datos reales de clientes ni código comentado.

Si todo está bien, aprueba y haz merge (o pide la aprobación de CODEOWNERS si toca archivos críticos).

### Paso 9 — ✋ Despliegue

El despliegue a producción siempre lo aprueba una persona. Sigue el procedimiento del proyecto y, después de desplegar, verifica que la funcionalidad responda y revisa los logs durante unos minutos.

### Resumen del flujo

| Paso | Quién | Tu acción | Tiempo típico |
|---|---|---|---|
| 1. Pedir | Tú | Describir la necesidad | 10 min |
| 2. Aclarar | Analista + tú | Responder preguntas | 10–30 min |
| 3. Spec | Analista | **Aprobar** | 15–30 min |
| 4. Plan | Arquitecto + UX | **Aprobar** | 20–45 min |
| 5. Tareas | Arquitecto + revisor | Vistazo rápido | 5 min |
| 6. Implementar | Desarrolladores | Supervisar | variable |
| 7. Validar | QA + revisor + seguridad | Intervenir si se atasca | variable |
| 8. PR | Orquestador | **Revisar y aprobar** | 20–40 min |
| 9. Desplegar | DevOps | **Aprobar** | 10 min |

---

## 6. Otros flujos: bugs y cambios pequeños

### Corregir un bug

> Usa la skill equipo-bug: al editar un pedido, la dirección de envío se borra.

El flujo es: diagnóstico de la causa → prueba que reproduce el bug → arreglo mínimo → validación → CHANGELOG. Revisa que el diagnóstico tenga sentido **antes** de que arregle.

### Cambios pequeños (sin el flujo completo)

Para cambios triviales (un texto, un color, un error de ortografía) no hace falta spec ni plan:

> Cambia el texto del botón "Enviar" por "Crear cuenta" en la pantalla de registro. Después usa la skill equipo-revision.

**Regla:** si el cambio afecta comportamiento, datos o API, **no es pequeño**. Usa el flujo completo.

### Retomar el trabajo (otra sesión, otro día u otra persona)

El estado no depende de la memoria del agente: vive en `docs/producto/roadmap.md` y en `specs/<rama>/estado.md`, versionados en git.

1. Cámbiate a la rama de la funcionalidad (`git checkout 003-registro-usuarios`) y ejecuta `make estado`. Verás la fase, las aprobaciones, las tareas hechas, los hallazgos abiertos y el próximo paso.
2. Abre el agente. El orquestador te dice dónde quedaron y pregunta si sigue. También puedes pedirlo:
   > ¿Por dónde quedamos?
3. Confirma que el resumen es correcto antes de que continúe. Si falta una aprobación registrada, te la va a pedir: léela antes de aprobar, no la des por hecha.

**Antes de dejar el trabajo**, dile al orquestador "lo dejamos por hoy": actualiza `estado.md` con el próximo paso y lo guarda en un commit. Si quedan cambios de código sin commit, te avisa.

Si una funcionalidad se empezó sin `estado.md` (por ejemplo, antes de que existiera este flujo), pide "retomemos": el orquestador lo reconstruye a partir de los archivos y de git, y te pide confirmar las aprobaciones.

### Costo de IA de cada funcionalidad

Cada funcionalidad lleva su cuenta en `specs/<rama>/costos.json`: cuántos tokens gastó cada agente con cada modelo y cuánto equivale en dólares.

- **Se registra solo.** El orquestador ejecuta `make costos` al iniciar y al cerrar cada sesión. Toma el consumo de las sesiones de OpenCode de tu computadora y lo guarda en la rama; como va en git, suma el trabajo de todas las personas que tocaron la funcionalidad.
- **El precio de cada día queda guardado.** Los precios salen de [models.dev](https://models.dev). Si un precio cambia, la tarea agrega la versión nueva con su fecha, sin recalcular lo anterior: cada respuesta se cobra con el precio vigente cuando ocurrió. La tarifa de hora pico de DeepSeek se aplica según la hora de cada respuesta.
- **Al aprobar el PR se cierra.** El orquestador ejecuta `make costos CERRAR=1` antes del merge. Desde ahí el costo de esa funcionalidad no cambia; el pre-commit bloquea cualquier modificación.
- **Es un costo equivalente.** Con OpenCode Go pagas una suscripción: el dólar es lo que costaría a precio de API. Tu gasto real es el % de la consola.

Comandos:

```bash
make costos              # tarea actual: tokens y $ por agente y modelo
make costos TODO=1       # todo el proyecto: por funcionalidad, agente y modelo
make costos PRECIOS=hoy  # cuánto costaría hoy (para cotizar; no guarda nada)
```

Si trabajaste en una sesión de OpenCode sin pasar por el orquestador, ejecuta `make costos` en la rama de la funcionalidad antes de cambiarte de rama: el consumo se asigna a la rama en la que estés al registrarlo.

### Cambiar la especificación de algo ya construido

Primero se actualiza `spec.md` y se aprueba; después se actualiza el plan y las tareas; después el código. **Nunca al revés.**

---

## 7. Reglas de oro

1. **Nunca apruebes lo que no leíste.** Tu aprobación es tu firma.
2. **Nunca uses `git commit --no-verify`** ni permitas que el agente lo haga.
3. **Nunca pegues secretos** (contraseñas, claves de API, datos de clientes) en el chat del agente ni en archivos del repositorio.
4. **Nunca edites archivos generados** (`CLAUDE.md`, `.claude/`, `.codex/`, `.opencode/`). Cambia la fuente y ejecuta `make sincronizar`.
5. **Nunca modifiques una migración ya aplicada.** Se crea una nueva.
6. **No aceptes pruebas desactivadas o debilitadas** para que "pase el CI".
7. **Ante la duda, pregunta**: al cliente si es de negocio, a un senior si es técnica.
8. **Si el agente se atasca o da vueltas**, detenlo. Una instrucción más precisa vale más que diez intentos.
9. **Reporta lo que se repite.** Si un error de los agentes aparece varias veces, avísale a dirección técnica para mejorar la skill o la constitución.

---

## 8. Problemas comunes

| Síntoma | Causa probable | Solución |
|---|---|---|
| `make doctor` dice "Hooks de git inactivos" | No se ejecutó la instalación | `make instalar-hooks` |
| "Submódulo .bowser-spec-kit-ai/ sin inicializar" o la carpeta `.bowser-spec-kit-ai/` está vacía | Se clonó sin `--recursive` | `git submodule update --init` |
| "Los archivos del kit no coinciden con .bowser-spec-kit-ai/" al hacer commit | Se editó a mano un archivo del kit, o se actualizó `.bowser-spec-kit-ai/` sin instalar | `make verificar-kit` para ver cuál; luego revierte el cambio o ejecuta `make instalar-kit` |
| `make actualizar-kit` se detiene por archivos modificados | Alguien cambió localmente un archivo del kit | Consulta a dirección técnica: llevarlo al kit, excluirlo en `equipo/config.json` o descartarlo con `FORZAR=1` |
| Actualicé el kit pero `.opencode/` (o los modelos) siguen iguales | Los modelos salen de `equipo/config.json` del proyecto, que el kit no sobrescribe | `make actualizar-modelos` y reinicia OpenCode |
| `make estado` avisa "cambió después de estado.md" o "aprobación no registrada" | El estado no se actualizó en la última sesión | Pide al orquestador "¿por dónde quedamos?": compara con los archivos, corrige `estado.md` y te pide confirmar lo que falte |
| ⚠ "sin actualizar estado.md" al hacer commit | Cambió spec, plan o tareas y el estado no | Es un aviso, no bloquea. Actualiza `estado.md` y agrégalo al commit |
| `make costos` dice "No se encontró 'opencode'" | La terminal no es la de Ubuntu/WSL, o OpenCode no está instalado ahí | Abre la terminal de Ubuntu (sección 3.1.1) |
| `make costos` dice "Aún no hay consumo registrado" aunque ya trabajaste | Kit anterior a la 1.6.3 con OpenCode 2.x, o las sesiones se hicieron en otra carpeta o con otra herramienta | `make actualizar-kit` y vuelve a ejecutar `make costos`; si avisa que OpenCode no devolvió sesiones, comprueba con `opencode session list` desde la carpeta del proyecto |
| `make costos` dice "Sin precio para: …" | El modelo no está en models.dev | Agrega su precio en `equipo/config.json` → `costos.precios_manuales` (ver `equipo/MODELOS.md`) |
| "El costo de … ya está cerrado" al hacer commit | Se modificó el `costos.json` de una funcionalidad terminada | Revierte el cambio (`git checkout -- <archivo>`). Solo dirección técnica puede autorizar una corrección |
| El CI falla en `golangci-lint` con "the Go language version (…) used to build golangci-lint is lower than the targeted Go version" | El proyecto usa un Go más nuevo que el golangci-lint que fija el kit | `make actualizar-kit`; si el kit aún no lo trae, crea la variable `GOLANGCI_LINT_VERSION` (ej. `v2.14.0`) en GitHub → Settings → Secrets and variables → Actions → Variables |
| El CI falla al descargar el submódulo | El repositorio del kit es privado | Configura el secreto `KIT_TOKEN` en el repositorio del proyecto |
| `make: command not found` | Falta `make` (común en Ubuntu/WSL) | `sudo apt install -y make build-essential` |
| `make: *** No rule to make target 'up'` | Estás en otra carpeta (por ejemplo, la terminal se abrió en tu carpeta personal) | `cd` a la carpeta del proyecto; `ls Makefile` debe encontrarlo |
| `make: docker: No such file or directory` | Docker no está instalado en Ubuntu, o falta la integración WSL | Docker Desktop con **WSL Integration → Ubuntu** activado (ver 3.1.1) y reabre la terminal |
| `make doctor` dice que falta `jq` | No está instalado | `sudo apt install -y jq` |
| El commit falla **desde VS Code** pero funciona desde la terminal, con varios ✗ a la vez (kit, constitución, agentes) | VS Code abrió el proyecto desde Windows y usa el Git de Windows, donde no hay `python3` | Instala la extensión **WSL** y abre el proyecto con `code .` desde Ubuntu; abajo a la izquierda debe decir "WSL: Ubuntu" |
| El agente no puede ejecutar `make`, `go` o `git`, o `which opencode` muestra `/mnt/c/...` | Ubuntu está usando la versión de Windows del agente | Instálalo en Ubuntu, verifica con `which -a` y vuelve a iniciar sesión (sección 3.1.2) |
| `make doctor` dice "Se están usando versiones de Windows de: …" | Esas herramientas no están instaladas en Ubuntu y se toman las de Windows | Instálalas en Ubuntu (sección 3.2) y verifica con `which -a <herramienta>` |
| `authentication method 10 not supported` | El cliente de base de datos es demasiado viejo para PostgreSQL 16 (por ejemplo Navicat 11) | Usa DBeaver Community o pgAdmin 4, o actualiza tu cliente |
| Error de autenticación **en español** ("la autentificación password falló…") | Te estás conectando a un PostgreSQL instalado en Windows, no al de Docker (el de Docker responde en inglés) | Desinstala o detén el PostgreSQL de Windows (`Get-Service *postgres*` en PowerShell), o cambia el puerto del de Docker |
| "Configuración de agentes desactualizada" | Alguien cambió `equipo/` o `.agents/` sin regenerar | `make sincronizar` y commit |
| El commit se rechaza por "Mensaje de commit inválido" | No sigue Conventional Commits | Usa `feat: …`, `fix: …`, `docs: …`, etc. |
| El commit se rechaza por la constitución | Se modificó `constitution.md` | Revierte el cambio; solo dirección técnica puede aprobarlo |
| El commit se rechaza por una migración | Se editó una migración existente | Revierte y crea una migración nueva |
| El agente no conoce los comandos `/speckit.*` | Spec Kit no está inicializado para esa herramienta | `specify init --here --force --integration <herramienta>` |
| El agente no usa los subagentes | Configuración no generada o herramienta sin soporte | `make sincronizar`; si no hay soporte, el orquestador asume los roles en secuencia |
| "Docker no está corriendo" | Docker Desktop cerrado | Ábrelo y espera a que arranque |
| Error de conexión a la base de datos | PostgreSQL no levantó o `.env` incorrecto | `make up`, `docker compose ps`, revisa `DATABASE_URL` en `.env` |
| La contraseña de `.env` no funciona | PostgreSQL solo toma la contraseña al crear la base por primera vez; se cambió `.env` después | `docker compose down -v && make up` (**borra los datos locales**) |
| El puerto 5432 está ocupado | Hay otro PostgreSQL local | Detén el otro, o pon otro puerto en `POSTGRES_PORT` dentro de `.env` (ej. `5433`) y ajusta `DATABASE_URL` |
| `golangci-lint` o `migrate`: "command not found" | `GOPATH/bin` no está en el `PATH` | Agrega `export PATH="$PATH:$(go env GOPATH)/bin"` a tu `~/.zshrc` o `~/.bashrc` |
| El agente da vueltas sin terminar una tarea | Instrucción ambigua o tarea muy grande | Detenlo, divide la tarea o da una instrucción concreta |

---

## 9. Glosario

- **Spec (especificación):** documento que describe qué se construye y por qué, con criterios de aceptación.
- **Plan:** documento que describe cómo se construye (arquitectura, datos, API).
- **Constitución:** reglas no negociables que todo el código debe cumplir.
- **Criterio de aceptación:** condición verificable que debe cumplirse para dar algo por terminado.
- **Orquestador:** el agente principal que coordina a los subagentes.
- **Subagente:** agente especializado en un rol (arquitecto, QA…).
- **Skill:** conjunto de instrucciones reutilizables (convenciones del stack o un flujo de trabajo).
- **Spec Kit:** herramienta de GitHub que estructura el proceso en fases (`specify`, `plan`, `tasks`, `implement`, `converge`).
- **Hook de git:** control automático que se ejecuta en cada commit.
- **CI:** integración continua; pruebas y controles automáticos en GitHub en cada PR.
- **PR (Pull Request):** solicitud para integrar cambios a la rama principal.
- **Migración:** archivo versionado que cambia el esquema de la base de datos.
- **Conventional Commits:** formato de mensajes de commit (`feat:`, `fix:`…).
- **Roadmap:** lista ordenada de funcionalidades del producto con el estado de cada una.
- **costos.json:** archivo de cada funcionalidad con los tokens y el costo equivalente de IA por agente y modelo, con los precios vigentes en cada fecha.
- **estado.md:** archivo de cada funcionalidad donde el orquestador anota fase, aprobaciones, hallazgos, decisiones y próximo paso, para retomar en cualquier sesión.

---

## 10. Lista de la primera semana

**Día 1**
- [ ] Recibir todos los accesos (sección 2).
- [ ] Instalar el entorno (sección 3).
- [ ] Clonar un proyecto, `make doctor` en verde (sección 4).
- [ ] Leer la constitución y `AGENTS.md`.

**Día 2**
- [ ] Leer las specs y planes de 2 funcionalidades ya terminadas, con su `estado.md`.
- [ ] Ejecutar `make estado` y entender el roadmap del proyecto.
- [ ] Leer las definiciones de los roles en `equipo/agentes/`.
- [ ] Hacer un cambio pequeño (sección 6) y abrir un PR acompañado de un senior.

**Días 3 a 5**
- [ ] Desarrollar una funcionalidad de práctica completa con el flujo de la sección 5, en una rama que no se integrará.
- [ ] Revisar tus aprobaciones de spec y plan con un senior.
- [ ] Anotar dudas o fricciones y compartirlas con dirección técnica: esta guía mejora con tu experiencia.
