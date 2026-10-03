# Comandos propios del proyecto (incluido por el Makefile del kit al final: no edites el Makefile).
# Nota: api-gen y e2e quedan definidos antes que su contenido (llegan con T022 y T027):
# se ejecutan cuando existan frontend/package.json (npm run api:gen) y frontend/e2e/playwright.config.ts.
# sqlc-verify valida su contenido en T006 (backend/sqlc.yaml).

.PHONY: api-gen sqlc-gen sqlc-verify e2e

api-gen: ## Regenera los tipos TypeScript desde el contrato OpenAPI del backend
	cd frontend && npm run api:gen

sqlc-gen: ## Genera el código Go de consultas desde backend/sqlc.yaml
	cd backend && sqlc generate

sqlc-verify: ## Verifica que el código generado por sqlc está al día (sin deriva)
	cd backend && sqlc generate
	git diff --exit-code -- backend/internal/db backend/sqlc.yaml

e2e: ## Ejecuta las pruebas end-to-end con Playwright (requiere `make up` levantado)
	cd frontend && npx playwright test --config e2e/playwright.config.ts
