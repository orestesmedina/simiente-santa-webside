# Imagen del backend (T019, plan D5).
#
# Build multi-stage: la única etapa de compilación parte de golang:1.27 (D5),
# con go.mod/go.sum copiados primero para cachear la descarga de dependencias,
# y produce un binario estático (CGO_ENABLED=0). La imagen de ejecución es
# distroless/static-debian12 (no es un FROM de compilación): no lleva shell,
# ni gestor de paquetes, ni toolchain de Go, ni fuentes — solo el binario,
# ca-certificates, tzdata y un usuario no-root (uid 65532, "nonroot").
# La etiqueta va fijada (nada de :latest) para builds reproducibles (reglas
# devops, constitución VII: configuración por variables de entorno).
FROM golang:1.27 AS build

WORKDIR /src

# go.mod y go.sum primero: esta capa solo se invalida cuando cambian las
# dependencias, así que `go mod download` se cachea entre builds normales.
COPY go.mod go.sum ./
RUN go mod download

# El resto del código (el build es Go puro: sin CGO ni gcc que requiera musl/debian aparte).
COPY . .

# Binario estático, sin debug/test info, sin dependencias de libc.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w" \
    -o /out/api ./cmd/api

# --- Imagen final de ejecución (mínima) ---
FROM gcr.io/distroless/static-debian12:latest

# Historia/etiquetas útiles para operar (no llevan secretos).
LABEL org.opencontainers.image.title="simiente-santa backend" \
      org.opencontainers.image.description="API de Simiente Santa (F1: GET /healthz)" \
      org.opencontainers.image.source="https://github.com/simiente-santa/simiente_santa" \
      org.opencontainers.image.licenses="proprietary"

WORKDIR /app

# Binario compilado. Distroless/static ya incluye ca-certificates.crt y
# tzdata, así que no hay que copiarlos.
COPY --from=build /out/api /api

# Distroless/static-debian12 trae de serie el usuario no-root "nonroot"
# (uid/gid 65532); lo fijamos para no ejecutar como root (defensa en
# profundidad, constitución §IV).
USER nonroot:nonroot

# Puerto que el servidor escucha por HTTP_PORT (8080 por defecto; D12/T007).
EXPOSE 8080

# Distroless no tiene shell: ENTRYPOINT exec-form, sin overlay comando.
ENTRYPOINT ["/api"]
