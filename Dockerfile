# ─────────────────────────────────────────────────────────────────────────────
# DSMES Backend — Multi-stage Dockerfile
#
# Stage 1 (builder): compiles the Go binary with all build dependencies.
# Stage 2 (runtime): minimal image — only the compiled binary + ca-certs.
#
# This two-stage approach produces an image < 20 MB instead of > 800 MB.
# ─────────────────────────────────────────────────────────────────────────────

# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

# Install git (needed for go mod download of private modules, if any).
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy dependency manifests first to leverage Docker layer cache.
# The go mod download layer only rebuilds when go.mod or go.sum changes.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy all source code.
COPY . .

# Compile the binary.
# -ldflags="-s -w"  strips debug info and DWARF to reduce binary size.
# CGO_ENABLED=0     builds a statically linked binary (no libc dependency).
# GOARCH=amd64      explicitly target amd64 (change for ARM VPS).
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/server ./cmd/api

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM scratch

# ca-certificates: required for HTTPS outbound requests (e.g. external APIs).
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# tzdata: required for Asia/Makassar timezone in GORM and logger.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the compiled binary from the builder stage.
COPY --from=builder /app/server /server

# Expose the application port (must match APP_PORT in .env).
EXPOSE 8080

# Use a non-root user for security (numeric UID avoids needing /etc/passwd).
USER 1000:1000

ENTRYPOINT ["/server"]
