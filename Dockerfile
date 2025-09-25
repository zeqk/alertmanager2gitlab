# Etapa 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copiar archivos go
# COPY go.mod go.sum ./
# RUN go mod download

COPY . .

# Compilar binario
RUN go build -o /alert-gitlab

# Etapa 2: Runtime ligero
FROM alpine:latest

WORKDIR /

# Copiar binario compilado
COPY --from=builder /alert-gitlab /alert-gitlab

# Variables de entorno (pueden ser redefinidas al correr el contenedor)
ENV GITLAB_TOKEN=""
ENV GITLAB_PROJECT_ID=""
ENV GITLAB_API_URL="https://gitlab.com/api/v4"

# Exponer puerto
EXPOSE 8080

# Ejecutar la app
CMD ["/alert-gitlab"]
