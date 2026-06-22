# syntax=docker/dockerfile:1

# ---- Stage 1: build the frontend ----
FROM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# ---- Stage 2: build the Go binary ----
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY main.go ./
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/registryui .

# ---- Stage 3: minimal runtime ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates \
 && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/registryui ./registryui
COPY --from=web   /web/dist       ./web/dist
# REGISTRIES is the deploy-time allow-list of selectable registries, given as
# comma-separated "Name=URL" pairs, e.g.:
#   -e REGISTRIES="Production=https://registry.example.com,Local=http://localhost:5000"
# When unset it falls back to a single entry from REGISTRY_URL.
# Set TLS_CERT_FILE and TLS_KEY_FILE (mount the certs) to serve HTTPS; when
# both are set the server listens on TLS only and plain HTTP is disabled.
ENV PORT=:8080 \
    REGISTRY_URL=http://localhost:5000 \
    REGISTRIES="" \
    STATIC_DIR=web/dist \
    TLS_CERT_FILE="" \
    TLS_KEY_FILE=""
EXPOSE 8080
USER app
ENTRYPOINT ["./registryui"]
