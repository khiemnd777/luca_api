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
  echo "docker daemon is not running. Please start Docker Desktop or dockerd first."
  exit 1
fi

mkdir -p \
  "$ROOT_DIR/tmp/observability/logs" \
  "$ROOT_DIR/tmp/observability/loki" \
  "$ROOT_DIR/tmp/observability/promtail" \
  "$ROOT_DIR/tmp/observability/grafana"

touch "$ROOT_DIR/tmp/observability/logs/luca_api.json.log"

"${COMPOSE_CMD[@]}" -f "$ROOT_DIR/docker-compose.observability.yml" up -d

echo "Observability stack is running."
echo "Grafana: http://127.0.0.1:3001 (admin/admin)"
echo "Loki health: http://127.0.0.1:3100/ready"
