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

Trabajas con un agente principal (el **orquestador**) que coordina a 10 especialistas:

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
- **Windows**: instala **WSL2 con Ubuntu** y trabaja siempre dentro de WSL. Los hooks y scripts son de bash.
  ```powershell
  wsl --install -d Ubuntu
  ```
  Desde aquí, todos los comandos se ejecutan en la terminal de Ubuntu.

### 3.2 Herramientas base

**macOS** (con [Homebrew](https://brew.sh)):
```bash
brew install git go node@22 python@3.12 jq make gh gitleaks golang-migrate
brew install --cask docker        # Docker Desktop; ábrelo una vez para que arranque
```

**Ubuntu / WSL2:**
```bash
sudo apt update && sudo apt install -y git jq make curl build-essential python3 python3-pip
# Go 1.23+ : https://go.dev/doc/install
# Node 22  : https://github.com/nvm-sh/nvm  →  nvm install 22
# Docker   : https://docs.docker.com/engine/install/ubuntu/  (o Docker Desktop con integración WSL)
# GitHub CLI: https://github.com/cli/cli/blob/trunk/docs/install_linux.md
```

**uv y Spec Kit** (todos los sistemas):
```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
uv tool install specify-cli
specify --version
```

**Herramientas de Go** (todos los sistemas):
```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```
Asegúrate de que `$(go env GOPATH)/bin` esté en tu `PATH`.

### 3.3 El agente de código

Instala **el que use el equipo** (revisa `"herramientas"` en `equipo/config.json`):

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

`make doctor` debe terminar con **"Todo listo para trabajar"**. Si marca problemas, revisa la [sección 8](#8-problemas-comunes).

Luego lee, en este orden:
1. `README.md` del proyecto.
2. `.specify/memory/constitution.md`.
3. `AGENTS.md`.
4. La carpeta `specs/`, empezando por la funcionalidad más reciente, para ver cómo se ha trabajado.

**Sobre la carpeta `.bowser-spec-kit-ai/`:** es el kit compartido de la empresa, incluido como submódulo de git. Sus archivos se copian a la raíz del proyecto con `make instalar-kit`. **No edites `.bowser-spec-kit-ai/` ni los archivos que vienen de él** (están listados en `.kit-manifest.json`): el hook de git y el CI lo detectan. Si algo del kit debería cambiar, propónlo a dirección técnica.

### Caso B: crear un proyecto nuevo

Solo lo hace quien tenga autorización. Sigue la sección **Instalación** del `README.md` del kit: instalar Spec Kit, `specify init`, `git submodule add … .bowser-spec-kit-ai` y `make -f .bowser-spec-kit-ai/Makefile instalar-kit`.

### Actualizar el kit en un proyecto

Cuando dirección técnica publique una mejora del kit:

```bash
make actualizar-kit
git add . && git commit -m "chore: actualiza kit de desarrollo"
```

Si el comando se detiene porque un archivo del kit fue modificado en el proyecto, no uses `FORZAR=1` sin consultar: ese cambio local podría ser importante.

---

## 5. Flujo de desarrollo de principio a fin

Usaremos un ejemplo: **"Los clientes pueden registrarse con email y contraseña."**

### Paso 0 — Antes de empezar

```bash
git checkout main && git pull
make up
make doctor
```
Abre tu agente en la carpeta del proyecto (`claude`, `codex` u `opencode`).

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
| El CI falla al descargar el submódulo | El repositorio del kit es privado | Configura el secreto `KIT_TOKEN` en el repositorio del proyecto |
| `make: command not found` | Falta `make` (común en Ubuntu/WSL) | `sudo apt install -y make build-essential` |
| "Configuración de agentes desactualizada" | Alguien cambió `equipo/` o `.agents/` sin regenerar | `make sincronizar` y commit |
| El commit se rechaza por "Mensaje de commit inválido" | No sigue Conventional Commits | Usa `feat: …`, `fix: …`, `docs: …`, etc. |
| El commit se rechaza por la constitución | Se modificó `constitution.md` | Revierte el cambio; solo dirección técnica puede aprobarlo |
| El commit se rechaza por una migración | Se editó una migración existente | Revierte y crea una migración nueva |
| El agente no conoce los comandos `/speckit.*` | Spec Kit no está inicializado para esa herramienta | `specify init --here --force --integration <herramienta>` |
| El agente no usa los subagentes | Configuración no generada o herramienta sin soporte | `make sincronizar`; si no hay soporte, el orquestador asume los roles en secuencia |
| "Docker no está corriendo" | Docker Desktop cerrado | Ábrelo y espera a que arranque |
| Error de conexión a la base de datos | PostgreSQL no levantó o `.env` incorrecto | `make up`, `docker compose ps`, revisa `DATABASE_URL` en `.env` |
| El puerto 5432 está ocupado | Hay otro PostgreSQL local | Detén el otro o cambia el puerto en `docker-compose.yml` y `.env` |
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

---

## 10. Lista de la primera semana

**Día 1**
- [ ] Recibir todos los accesos (sección 2).
- [ ] Instalar el entorno (sección 3).
- [ ] Clonar un proyecto, `make doctor` en verde (sección 4).
- [ ] Leer la constitución y `AGENTS.md`.

**Día 2**
- [ ] Leer las specs y planes de 2 funcionalidades ya terminadas.
- [ ] Leer las definiciones de los roles en `equipo/agentes/`.
- [ ] Hacer un cambio pequeño (sección 6) y abrir un PR acompañado de un senior.

**Días 3 a 5**
- [ ] Desarrollar una funcionalidad de práctica completa con el flujo de la sección 5, en una rama que no se integrará.
- [ ] Revisar tus aprobaciones de spec y plan con un senior.
- [ ] Anotar dudas o fricciones y compartirlas con dirección técnica: esta guía mejora con tu experiencia.
