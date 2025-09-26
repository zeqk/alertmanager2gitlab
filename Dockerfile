# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy Go files
# COPY go.mod go.sum ./
# RUN go mod download

COPY . .

# Build binary
RUN go build -o /alertmanager2gitlab

# Stage 2: Lightweight runtime
FROM alpine:latest

WORKDIR /

# Copy compiled binary
COPY --from=builder /alertmanager2gitlab /alertmanager2gitlab
COPY --from=builder /app/templates /templates

# Environment variables (can be overridden when running the container)
ENV GITLAB_TOKEN=""
ENV GITLAB_PROJECT_ID=""
ENV GITLAB_API_URL="https://gitlab.com/api/v4"

# Expose port
EXPOSE 8080

# Run the app
CMD ["/alertmanager2gitlab"]
