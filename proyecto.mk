# Comandos propios del proyecto (incluido por el Makefile del kit al final: no edites el Makefile).
# Nota: api-gen y e2e quedan definidos antes que su contenido (llegan con T022 y T027):
# se ejecutan cuando existan frontend/package.json (npm run api:gen) y frontend/e2e/playwright.config.ts.
# sqlc-verify valida su contenido en T006 (backend/sqlc.yaml). R5: con
# internal/db/queries/ sin *.sql (F1) no hay nada que regenerar; sqlc exige al
# menos una consulta, así que el target lo detecta y termina en verde.

.PHONY: api-gen sqlc-gen sqlc-verify e2e

api-gen: ## Regenera los tipos TypeScript desde el contrato OpenAPI del backend
	cd frontend && npm run api:gen

sqlc-gen: ## Genera el código Go de consultas desde backend/sqlc.yaml
	cd backend && sqlc generate

sqlc-verify: ## Verifica que el código generado por sqlc está al día (sin deriva)
	@if [ -n "$$(ls backend/internal/db/queries/*.sql 2>/dev/null)" ]; then \
		$(MAKE) sqlc-gen; \
		git diff --exit-code -- backend/internal/db backend/sqlc.yaml; \
	else \
		echo "sqlc-verify: sin consultas en backend/internal/db/queries (R5 en F1): nada que regenerar"; \
	fi

e2e: ## Ejecuta las pruebas end-to-end con Playwright (requiere `make up` levantado)
	cd frontend && npx playwright test --config e2e/playwright.config.ts
