#!/bin/bash

echo "🚀 Boosting up..."

# Mặc định APP_ENV=development
ENV="development"

# Parse flag --env=...
for arg in "$@"; do
  case $arg in
    --env=*)
      ENV="${arg#*=}"
      shift
      ;;
  esac
done

echo "🌱 APP_ENV=$ENV"

# Chạy Go với biến môi trường APP_ENV
APP_ENV=$ENV go run main.go
