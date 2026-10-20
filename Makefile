# Makefile del kit. En los proyectos lo gestiona `make install-kit`: no lo edites ahí.
# Para agregar comandos propios de un proyecto, créalos en proyecto.mk (se incluye al final).

.PHONY: help status costs news doctor profile skills check-ci upgrade models update-models sync check-agents install-hooks install-kit update-kit check-kit up down generate check-generated test coverage lint security ci

# Carpeta del submódulo del kit: la del `make -f <carpeta>/Makefile` usado, o la guardada al instalar,
# o el nombre por defecto. Se puede forzar con `make ... KIT=<carpeta>`.
KIT_INVOCADO := $(patsubst %/,%,$(filter-out ./,$(dir $(firstword $(MAKEFILE_LIST)))))
KIT_GUARDADO := $(shell bash scripts/ruta-kit.sh 2>/dev/null)
KIT ?= $(or $(KIT_INVOCADO),$(KIT_GUARDADO),.bowser-spec-kit-ai)

# Opciones de los comandos. Los nombres anteriores (en español) se siguen aceptando.
FORCE ?= $(FORZAR)
SINCE ?= $(DESDE)
ALL ?= $(TODO)
CLOSE ?= $(CERRAR)
PRICES ?= $(PRECIOS)
YES ?= $(SI)
DETECT ?= $(DETECTAR)
PART ?= $(PARTE)

help: ## Muestra los comandos disponibles
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

install-kit: ## Copia el kit (submódulo) a la raíz, regenera agentes y activa hooks
	@test -f $(KIT)/scripts/instalar_kit.py || { echo "No existe $(KIT)/. Agrega el submódulo (git submodule add <url> $(KIT)) o ejecuta: git submodule update --init"; exit 1; }
	@PREVIA=$$(python3 $(KIT)/scripts/instalar_kit.py --version-instalada) && \
	NOVEDADES_AL_FINAL=1 python3 $(KIT)/scripts/instalar_kit.py $(if $(FORCE),--forzar,) && \
	python3 scripts/sincronizar.py && \
	git config core.hooksPath .githooks && \
	{ python3 scripts/actualizar_modelos.py --comprobar || true; } && \
	python3 $(KIT)/scripts/instalar_kit.py --resumen-novedades "$$PREVIA" && \
	echo "" && echo "Listo. Revisa los cambios con 'git status' y haz commit (incluye .kit-manifest.json)."

update-kit: ## Trae la última versión del kit y la instala
	git submodule update --init --remote $(KIT)
	@$(MAKE) --no-print-directory install-kit

news: ## Historial de cambios del kit (SINCE=1.4.0 para ver desde una versión)
	@python3 $(KIT)/scripts/instalar_kit.py --novedades $(if $(SINCE),--desde $(SINCE),)

check-kit: ## Comprueba que la raíz coincida con la versión del submódulo del kit
	python3 $(KIT)/scripts/instalar_kit.py --verificar

status: ## Por dónde vamos: roadmap, fase, aprobaciones, tareas y próximo paso
	@python3 scripts/estado.py $(if $(ALL),--todo,)

costs: ## Costo de IA de la tarea actual (ALL=1 proyecto, CLOSE=1 cierra, PRICES=today cotiza)
	@python3 scripts/costos.py $(if $(ALL),--todo,) $(if $(CLOSE),--cerrar,) $(if $(filter today hoy,$(PRICES)),--hoy,)

doctor: ## Verifica que el entorno tenga todo lo necesario
	@bash scripts/doctor.sh

sync: ## Genera la configuración de Claude Code, Codex y OpenCode desde equipo/
	python3 scripts/sincronizar.py

models: ## Muestra qué modelo usa cada agente en cada herramienta
	@python3 scripts/sincronizar.py --modelos

update-models: ## Aplica al proyecto los modelos recomendados por el kit (muestra los cambios antes)
	@python3 scripts/actualizar_modelos.py $(if $(YES),--si,)

upgrade: ## Qué le falta al proyecto para aprovechar el kit (2.0), y arregla lo mecánico (YES=1 sin preguntar)
	@python3 scripts/pasar_a_2.py $(if $(YES),--si,)

check-agents: ## Comprueba que la configuración generada esté al día
	python3 scripts/sincronizar.py --verificar

install-hooks: ## Activa los hooks de git del proyecto (una vez por clon)
	git config core.hooksPath .githooks
	chmod +x .githooks/* equipo/adaptadores/claude/hooks/*.sh
	@echo "Hooks de git activados."

# El entorno local es del proyecto: el kit no trae ningún servicio. Lo crea el rol devops según el perfil.
COMPOSE := $(wildcard compose.yaml compose.yml docker-compose.yaml docker-compose.yml)
SIN_COMPOSE = echo "Este proyecto todavía no tiene entorno local: falta compose.yaml (o docker-compose.yml) en la raíz. Lo crea el rol devops con los servicios que el proyecto usa."; exit 1

up: ## Levanta el entorno local del proyecto (sus servicios, con Docker)
ifdef COMPOSE
	docker compose up -d
else
	@$(SIN_COMPOSE)
endif

down: ## Detiene el entorno local (conserva los datos)
ifdef COMPOSE
	docker compose down
else
	@$(SIN_COMPOSE)
endif

profile: ## Muestra y valida el perfil del proyecto (DETECT=1: qué tecnologías encuentra en el proyecto)
	@python3 scripts/perfil.py $(if $(DETECT),--detectar,)

skills: ## Catálogo de skills de tecnología del kit: madurez de cada una y cuáles tiene el proyecto
	@python3 scripts/catalogo.py

check-ci: ## Qué verbos del perfil ejecuta la integración continua del proyecto, y cuáles no ejecuta nadie
	@python3 scripts/ci.py

# Cada comando ejecuta el verbo que el perfil (equipo/perfil.json) declara para cada parte del proyecto
# (PART=nombre lo limita a una). Sin perfil, no hay nada que ejecutar: piden crearlo.
VERBO = python3 scripts/verbos.py $(1) $(if $(PART),--parte $(PART),)

generate: ## Regenera el código generado
	@$(call VERBO,generar)

check-generated: ## Regenera y falla si el código generado no estaba al día
	@$(call VERBO,generar --verificar)

test: ## Ejecuta todas las pruebas
	@$(call VERBO,probar)

coverage: ## Cobertura de pruebas (falla bajo el mínimo del proyecto)
	@$(call VERBO,cobertura)

lint: ## Formato y análisis estático
	@$(call VERBO,formato revisar)

security: ## Auditoría de dependencias
	@$(call VERBO,auditar)

ci: ## Lo mismo que corre en CI, con un resumen de lo comprobado
	@$(call VERBO,formato revisar generar probar cobertura auditar --verificar --resumen)

# Nombres anteriores de los comandos (en español). Siguen funcionando: ejecutan el comando nuevo y avisan de
# su nombre. Se retiran en una versión posterior a la 2.0.
define ANTERIOR
.PHONY: $(1)
$(1): $(2)
	@echo "" >&2; echo "Aviso: 'make $(1)' ahora se llama 'make $(2)'. El nombre anterior se retirará en una versión futura." >&2
endef
ANTERIORES := estado:status costos:costs novedades:news modelos:models actualizar-kit:update-kit 	instalar-kit:install-kit verificar-kit:check-kit actualizar-modelos:update-models sincronizar:sync 	verificar-agentes:check-agents instalar-hooks:install-hooks generar:generate 	verificar-generados:check-generated cobertura:coverage
$(foreach par,$(ANTERIORES),$(eval $(call ANTERIOR,$(word 1,$(subst :, ,$(par))),$(word 2,$(subst :, ,$(par))))))

# Comandos propios del proyecto (opcional; no lo gestiona el kit).
-include proyecto.mk

# Comandos para mantener el kit (solo existe en el repositorio del kit; no se copia a los proyectos).
-include kit.mk
