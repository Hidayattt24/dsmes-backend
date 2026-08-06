#!/bin/sh
# ─────────────────────────────────────────────────────────────────────────────
# DSMES Backend — container entrypoint
#
# Runs database migrations (idempotent) before starting the API server, so a
# fresh container is always migrated before it accepts traffic.
# ─────────────────────────────────────────────────────────────────────────────
set -e

echo "[entrypoint] running database migrations..."
/app/migrate

echo "[entrypoint] starting API server..."
exec /app/server
