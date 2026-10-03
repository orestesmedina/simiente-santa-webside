# Makefile del kit. En los proyectos lo gestiona `make instalar-kit`: no lo edites ahí.
# Para agregar comandos propios de un proyecto, créalos en proyecto.mk (se incluye al final).

.PHONY: help estado doctor modelos actualizar-modelos sincronizar verificar-agentes instalar-hooks instalar-kit actualizar-kit verificar-kit up down db-migrate test test-backend test-frontend lint security ci

# Carpeta del submódulo del kit: la del `make -f <carpeta>/Makefile` usado, o la guardada al instalar,
# o el nombre por defecto. Se puede forzar con `make ... KIT=<carpeta>`.
KIT_INVOCADO := $(patsubst %/,%,$(filter-out ./,$(dir $(firstword $(MAKEFILE_LIST)))))
KIT_GUARDADO := $(shell bash scripts/ruta-kit.sh 2>/dev/null)
KIT ?= $(or $(KIT_INVOCADO),$(KIT_GUARDADO),.bowser-spec-kit-ai)
FORZAR ?=

help: ## Muestra los comandos disponibles
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

instalar-kit: ## Copia el kit (submódulo) a la raíz, regenera agentes y activa hooks
	@test -f $(KIT)/scripts/instalar_kit.py || { echo "No existe $(KIT)/. Agrega el submódulo (git submodule add <url> $(KIT)) o ejecuta: git submodule update --init"; exit 1; }
	python3 $(KIT)/scripts/instalar_kit.py $(if $(FORZAR),--forzar,)
	python3 scripts/sincronizar.py
	git config core.hooksPath .githooks
	@python3 scripts/actualizar_modelos.py --comprobar || true
	@echo "Listo. Revisa los cambios con 'git status' y haz commit (incluye .kit-manifest.json)."

actualizar-kit: ## Trae la última versión del kit y la instala
	git submodule update --init --remote $(KIT)
	@$(MAKE) --no-print-directory instalar-kit

verificar-kit: ## Comprueba que la raíz coincida con la versión del submódulo del kit
	python3 $(KIT)/scripts/instalar_kit.py --verificar

estado: ## Por dónde vamos: roadmap, fase, aprobaciones, tareas y próximo paso
	@python3 scripts/estado.py $(if $(TODO),--todo,)

doctor: ## Verifica que el entorno tenga todo lo necesario
	@bash scripts/doctor.sh

sincronizar: ## Genera la configuración de Claude Code, Codex y OpenCode desde equipo/
	python3 scripts/sincronizar.py

modelos: ## Muestra qué modelo usa cada agente en cada herramienta
	@python3 scripts/sincronizar.py --modelos

actualizar-modelos: ## Aplica al proyecto los modelos recomendados por el kit (muestra los cambios antes)
	@python3 scripts/actualizar_modelos.py $(if $(SI),--si,)

verificar-agentes: ## Comprueba que la configuración generada esté al día
	python3 scripts/sincronizar.py --verificar

instalar-hooks: ## Activa los hooks de git del proyecto (una vez por clon)
	git config core.hooksPath .githooks
	chmod +x .githooks/*
	@echo "Hooks de git activados."

up: ## Levanta el entorno local (PostgreSQL y servicios)
	docker compose up -d

down: ## Detiene el entorno local (conserva los datos)
	docker compose down

db-migrate: ## Aplica las migraciones pendientes
	migrate -path backend/migrations -database "$$DATABASE_URL" up

test: test-backend test-frontend ## Ejecuta todas las pruebas

test-backend: ## Pruebas del backend (unitarias + integración)
	cd backend && go test ./... && go test -tags=integration ./...

test-frontend: ## Pruebas del frontend
	cd frontend && npm test -- --run

lint: ## Linters de backend y frontend
	cd backend && gofmt -l . && go vet ./... && golangci-lint run
	cd frontend && npm run lint && npm run typecheck

security: ## Auditoría de dependencias
	cd backend && govulncheck ./...
	cd frontend && npm audit --audit-level=high

ci: lint test security ## Lo mismo que corre en CI

# Comandos propios del proyecto (opcional; no lo gestiona el kit).
-include proyecto.mk
