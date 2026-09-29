# Dockerfile multietapa de la API de libros.
#
# IMPORTANTE: el CONTEXTO de construcción es la RAÍZ de demos-sesion4
# (ahí está go.mod), y este archivo se indica con -f:
#
#   cd demos-sesion4
#   docker build -f api/Dockerfile -t api-libros:v1 .
#
# ---------- Etapa 1: compilar ----------
FROM golang:1.26-alpine AS compilador
WORKDIR /src
# Primero solo go.mod: si no cambia, Docker reutiliza esta capa (caché).
COPY go.mod ./
COPY api/ ./api/
# CGO_ENABLED=0 -> binario estático (no necesita libc en la imagen final).
# -ldflags "-s -w" -> quita tablas de símbolos y depuración (binario más chico).
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /out/api ./api

# ---------- Etapa 2: imagen final mínima ----------
# distroless/static: sin shell, sin gestor de paquetes; solo lo mínimo.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=compilador /out/api /api
# Usuario NUMÉRICO: runAsNonRoot de Kubernetes exige número, no nombre.
USER 65532:65532
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/api"]
