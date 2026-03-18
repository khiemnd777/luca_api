#!/bin/bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

if docker compose version >/dev/null 2>&1; then
  COMPOSE_CMD=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE_CMD=(docker-compose)
else
  echo "docker compose is required but not found."
  exit 1
fi

if ! docker info >/dev/null 2>&1; then
  echo "docker daemon is not running. Nothing to stop."
  exit 0
fi

"${COMPOSE_CMD[@]}" -f "$ROOT_DIR/docker-compose.observability.yml" down

echo "Observability stack stopped."
