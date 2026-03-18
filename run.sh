#!/bin/bash
set -euo pipefail

echo "🚀 Boosting up..."

# Mặc định APP_ENV=development
ENV="development"
OBSERVABLE="false"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GOCACHE_DIR="$ROOT_DIR/.gocache"
LOG_DIR="$ROOT_DIR/tmp/observability/logs"
LOG_FILE="$LOG_DIR/luca_api.json.log"

for arg in "$@"; do
  case $arg in
    --env=*)
      ENV="${arg#*=}"
      shift
      ;;
    --observable)
      OBSERVABLE="true"
      shift
      ;;
  esac
done

echo "🌱 APP_ENV=$ENV"

mkdir -p "$GOCACHE_DIR"

if [ "$OBSERVABLE" = "true" ]; then
  echo "🧱 Starting local observability stack"
  "$ROOT_DIR/observability_up.sh"
  mkdir -p "$LOG_DIR"
  touch "$LOG_FILE"
  echo "📡 Observability log shipping enabled"
  echo "📝 Mirroring stdout/stderr to $LOG_FILE"
  APP_ENV="$ENV" GOFLAGS=-mod=mod GOCACHE="$GOCACHE_DIR" go run main.go 2>&1 | tee -a "$LOG_FILE"
  exit ${PIPESTATUS[0]}
fi

APP_ENV="$ENV" GOFLAGS=-mod=mod GOCACHE="$GOCACHE_DIR" go run main.go
