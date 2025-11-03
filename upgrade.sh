#!/bin/bash

set -e

LEVELS=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --patch) LEVELS+=("patch"); shift ;;
    --minor) LEVELS+=("minor"); shift ;;
    --major) LEVELS+=("major"); shift ;;
    *) echo "❌ Unknown option: $1"; exit 1 ;;
  esac
done

if [ ${#LEVELS[@]} -eq 0 ]; then
  LEVELS=("patch")
fi

echo "🔍 Upgrade level(s):"
for level in "${LEVELS[@]}"; do
  echo "  - $level"
done

echo "📦 Checking for outdated modules..."
OUTDATED=$(go list -u -m -mod=mod all 2>/dev/null | grep '\[')
if [[ -z "$OUTDATED" ]]; then
  echo "✅ All modules are up to date."
  exit 0
fi

TO_UPGRADE=()

while read -r line; do
  mod_path=$(echo "$line" | awk '{print $1}')
  current_ver=$(echo "$line" | awk '{print $2}')
  update_ver=$(echo "$line" | grep -o '\[.*\]' | tr -d '[]')

  current_parts=(${current_ver//./ })
  update_parts=(${update_ver//./ })

  match=false
  if [[ " ${LEVELS[@]} " =~ " major " ]] && [[ "${current_parts[0]}" != "${update_parts[0]}" ]]; then
    match=true
  elif [[ " ${LEVELS[@]} " =~ " minor " ]] && [[ "${current_parts[1]}" != "${update_parts[1]}" ]]; then
    match=true
  elif [[ " ${LEVELS[@]} " =~ " patch " ]] && [[ "${current_parts[2]}" != "${update_parts[2]}" ]]; then
    match=true
  fi

  if [ "$match" = true ]; then
    TO_UPGRADE+=("$mod_path $current_ver $update_ver")
  fi
done <<< "$OUTDATED"

if [ ${#TO_UPGRADE[@]} -eq 0 ]; then
  echo "✅ No modules match selected upgrade levels."
  exit 0
fi

echo ""
echo "📝 The following modules will be upgraded:"
for entry in "${TO_UPGRADE[@]}"; do
  mod=$(echo "$entry" | awk '{print $1}')
  from=$(echo "$entry" | awk '{print $2}')
  to=$(echo "$entry" | awk '{print $3}')
  echo "  - $mod $from → $to"
done

echo ""
read -p "❓ Proceed with upgrade? (y/N): " confirm
confirm=$(echo "$confirm" | tr '[:upper:]' '[:lower:]')
if [[ "$confirm" != "y" && "$confirm" != "yes" ]]; then
  echo "❌ Aborted."
  exit 0
fi

echo ""
for entry in "${TO_UPGRADE[@]}"; do
  mod=$(echo "$entry" | awk '{print $1}')
  old=$(echo "$entry" | awk '{print $2}')
  expected=$(echo "$entry" | awk '{print $3}')
  
  echo "🔼 go get $mod"
  if go get "$mod" &>/dev/null; then
    actual=$(go list -m all | grep "^$mod " | awk '{print $2}')
    if [[ "$actual" == "$expected" ]]; then
      echo "✅ Upgraded $mod $old → $actual"
    else
      echo "⚠️  Version mismatch: expected $expected but got $actual"
    fi
  else
    echo "❌ Failed to upgrade $mod"
  fi
  echo ""
done

echo ""
echo "🧹 Running go mod tidy..."
go mod tidy

echo "🔄 Syncing vendor directory with go mod vendor..."
go mod vendor

echo "✅ Modules upgraded and vendor directory synced."
echo "🏁 Done."
