# Imagen del frontend (T025, plan R6).
#
# Build multi-stage: en la primera etapa node:22 compila la SPA con npm ci &&
# npm run build (Vite) con VITE_API_URL horneada en build time vía ARG
# (plan R6: si cambia la URL de la API, hay que reconstruir la imagen).
# La segunda etapa es la imagen final: nginx:alpine sirviendo dist/ como estático
# con fallback de SPA (ver nginx.conf) — sin Node, ni node_modules, ni fuentes.
#
# Ambas imágenes van fijadas a una versión concreta (nada de :latest) para
# reproducción determinista; node:22 es la versión que fija el ci.yml del kit
# (plan R11/Node 22) en alpine para reducir el tamaño de la imagen.
FROM node:22.22-alpine AS build

# Values horneados en el build: VITE_API_URL, base URL de la API que el cliente
# (src/api/client.ts) consume desde import.meta.env. Por defecto apunta al
# backend local (plan R6): http://localhost:8080.
ARG VITE_API_URL=http://localhost:8080
ENV VITE_API_URL=${VITE_API_URL}

WORKDIR /app

# package.json y package-lock.json primero: la capa de npm ci solo se invalida
# cuando cambian las dependencias y se cachea entre build normales.
COPY package.json package-lock.json ./
RUN npm ci

# El resto del código (vite build genera dist/). La variable VITE_API_URL ya
# está disponible para import.meta.env porque fue declarada con ENV arriba.
COPY . .
RUN npm run build

# --- Imagen final: nginx sirviendo la build ---
FROM nginx:1.30.5-alpine

# nginx.conf reemplaza el server por defecto (puerto 80) con fallback de SPA
# y cabeceras de caché para index.html (sin cache) y /assets/ (inmutable).
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
