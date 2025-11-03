#!/bin/bash
set -e

VERSION_FILE="version.yaml"

# ✅ Nếu version.yaml đã được staged → skip
if git diff --cached --name-only | grep -q "^${VERSION_FILE}$"; then
  echo "⏭  $VERSION_FILE already staged. Skipping version bump."
  exit 0
fi

# 🧹 Clean macOS temp file
[ -f "${VERSION_FILE}-e" ] && rm -f "${VERSION_FILE}-e"

# 🔁 Bump logic
LINE=$(grep "^version:" "$VERSION_FILE")
BASE_VERSION=$(echo "$LINE" | cut -d '+' -f 1 | cut -d ' ' -f 2)
BUILD_VERSION=$(echo "$LINE" | cut -d '+' -f 2)
[ -z "$BUILD_VERSION" ] && BUILD_VERSION=0

NEW_BUILD=$((BUILD_VERSION + 1))
NEW_LINE="version: ${BASE_VERSION}+${NEW_BUILD}"

sed -i '' -e "s/^version:.*/$NEW_LINE/" "$VERSION_FILE"
git add "$VERSION_FILE"

echo "✅ Bumped version to $NEW_LINE (on develop)"
