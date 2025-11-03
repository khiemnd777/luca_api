#!/bin/bash

# ✅ Default path
APP_PATH="~/honvang_app/develop"

# ✅ Parse arguments
for arg in "$@"
do
  case $arg in
    --path=*)
      APP_PATH="${arg#*=}"
      shift
      ;;
    *)
      ;;
  esac
done

echo "🚀 Starting app from: $APP_PATH"

# ✅ Add Go to PATH
export PATH=$PATH:/usr/local/go/bin

# ✅ Change dir (expand ~ if needed)
cd $(eval echo "$APP_PATH") || {
  echo "❌ Cannot cd into $APP_PATH"
  exit 1
}

# ✅ Kill any old instance
pkill -f "go run ./main.go" || true

# ✅ Start app
setsid nohup go run ./main.go > ./dev.log 2>&1 < /dev/null &

echo "✅ App started. Logs: $APP_PATH/dev.log"
