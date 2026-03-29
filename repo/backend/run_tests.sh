#!/usr/bin/env sh
set -eu

echo "[backend] running go tests"
cd /app

export BASE_URL="https://localhost:${PORT:-8443}"
export ALLOW_INSECURE_TLS=true

cd /app/backend
go test ./internal/api/middleware ./internal/logger -v

cd /app
go test ./tests/unit_tests ./tests/API_tests -v
