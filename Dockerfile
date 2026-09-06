# ─────────────────────────────────────────────────────────────────────────────
# DSMES Backend — Multi-stage Dockerfile
#
# Stage 1 (builder):  compiles the server + migrate + worker binaries, generates
#                     the Swagger spec, and keeps the migration files.
# Stage 2 (runtime):  slim Alpine image with all three binaries + migrations + docs.
#
# The API container runs migrations automatically on start (see entrypoint.sh).
# The worker container overrides the entrypoint to run /app/worker directly, so
# it never executes migrations.
# ─────────────────────────────────────────────────────────────────────────────

# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

# git: needed for `go mod download`; swag: generates the API docs.
RUN apk add --no-cache git ca-certificates tzdata \
    && go install github.com/swaggo/swag/cmd/swag@v1.16.6

WORKDIR /app

# Copy dependency manifests first to leverage Docker layer cache.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy all source code.
COPY . .

# Generate Swagger documentation into ./docs (served by the runtime).
RUN swag init -g cmd/api/main.go -o ./docs --parseDependency --parseInternal

# Compile the binaries (static, stripped).
# CGO_ENABLED=0 → statically linked; GOARCH=amd64 (change for ARM VPS).
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/server ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/migrate ./cmd/migrate \
 && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/worker ./cmd/worker

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 1000 app && adduser -S -u 1000 -G app app

WORKDIR /app

# Binaries
COPY --from=builder /app/server  /app/server
COPY --from=builder /app/migrate /app/migrate
COPY --from=builder /app/worker  /app/worker

# Runtime data the server needs (migrations for the entrypoint, docs for Swagger)
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/docs       /app/docs

COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENV TZ=Asia/Jakarta

# Expose the application port (must match APP_PORT in .env).
EXPOSE 8080

# Use a non-root user for security.
USER 1000:1000

ENTRYPOINT ["/app/entrypoint.sh"]
