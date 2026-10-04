---
name: go-backend
description: Convenciones de la empresa para escribir backend en Go (estructura, capas, errores, HTTP, pruebas). Usar siempre que se cree o modifique código en backend/.
---
# Backend en Go — convenciones

## Versión y herramientas
- Go 1.26 o superior (una versión con soporte vigente). Router: `net/http` estándar (con patrones `GET /users/{id}`) o `chi` si el plan lo justifica.
- Base de datos: `pgx/v5` + `sqlc` para generar consultas tipadas. Migraciones: `golang-migrate`.
- Logs: `log/slog` en JSON. Configuración: variables de entorno.
- Lint: `golangci-lint`. Seguridad: `govulncheck`.

## Estructura
```
backend/
├── cmd/api/main.go          # arranque: config, DB, router, servidor
├── internal/
│   ├── config/              # carga de variables de entorno
│   ├── <dominio>/           # ej. users/, orders/
│   │   ├── handler.go       # HTTP: decodifica, valida, llama al service, responde
│   │   ├── service.go       # reglas de negocio (sin HTTP, sin SQL)
│   │   ├── repository.go    # acceso a datos (interfaz + implementación pgx/sqlc)
│   │   ├── model.go         # tipos del dominio y DTOs
│   │   └── *_test.go
│   ├── platform/            # db, http middleware, errores comunes
│   └── db/queries/*.sql     # consultas para sqlc
├── migrations/              # 000001_create_users.up.sql / .down.sql
└── api/openapi.yaml
```

## Reglas
- `handler` depende de una **interfaz** del `service`; `service` depende de una **interfaz** del `repository`. Esto permite probar cada capa aislada.
- Pasa `context.Context` como primer parámetro en todo lo que haga I/O.
- Errores: envuelve con contexto `fmt.Errorf("create user: %w", err)`. Define errores de dominio (`ErrNotFound`, `ErrConflict`) y tradúcelos a HTTP solo en el handler.
- Nunca expongas errores internos al cliente. Respuesta de error estándar:
  ```json
  { "error": { "code": "not_found", "message": "Usuario no encontrado" } }
  ```
- Valida toda entrada en el handler (campos requeridos, longitudes, formatos) y responde 400 con detalle.
- Cierra siempre recursos (`defer rows.Close()`), respeta cancelación de contexto, evita goroutines sin control.
- Timeouts en el servidor HTTP (`ReadHeaderTimeout`, `WriteTimeout`) y apagado ordenado con `signal.NotifyContext`.

## Pruebas
- Tabla de casos (`tests := []struct{...}`) con `t.Run`.
- `service`: pruebas unitarias con repositorio falso (fake) implementando la interfaz.
- `handler`: `httptest.NewRecorder` + service falso.
- `repository`: pruebas de integración contra PostgreSQL real (variable `DATABASE_URL_TEST`), marcadas con `//go:build integration`.
- Comandos: `go test ./...` (unitarias) y `go test -tags=integration ./...` (integración).
