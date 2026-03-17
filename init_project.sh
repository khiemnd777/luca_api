#!/bin/bash
set -e

export GOCACHE="${PWD}/.gocache"

echo "🚀 Initializing project..."

# Step 0: Delete vendor folder
if [ -d "vendor" ]; then
  echo "🧹 Removing vendor folder..."
  rm -rf vendor
fi

# Step 1: Generate Ent for shared
echo "👉 Generating Ent for shared"
go run -mod=mod entgo.io/ent/cmd/ent generate ./shared/db/ent/schema --target ./shared/db/ent/generated --feature sql/execquery

# Step 1.1: Init database
echo "👉 Initializing database"
GOFLAGS=-mod=mod go run scripts/init_db/main.go

# Step 2: Auto generate Ent for all modules with ent/schema
for schema_dir in $(find modules -type d -path "*/ent/schema"); do
  module_dir=$(dirname "$(dirname "$schema_dir")")
  schema_path="./$schema_dir"
  target_path="./$module_dir/ent/generated"

  echo "👉 Generating Ent for $module_dir"
  go run -mod=mod entgo.io/ent/cmd/ent generate "$schema_path" --target "$target_path"

  echo "⚙️ Running auto-migrate for $module_dir"
  (cd "$module_dir" && GOFLAGS=-mod=mod go run ./ent/cmd/migrate.go)
done

# Step 3: Tidy & Vendor
echo "👉 Running go mod tidy"
go mod tidy

echo "👉 Running go mod vendor"
go mod vendor

# Step 4: Init roles
echo "👉 Initializing roles"
GOFLAGS=-mod=mod go run scripts/init_roles/main.go

# Step 5: Build all
echo "👉 Building all modules"
GOFLAGS=-mod=mod go build ./...

echo "✅ Project initialized successfully!"
