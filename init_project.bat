@echo off
echo 🚀 Initializing project...

REM Step 0: Delete vendor folder
IF EXIST vendor (
    echo 🧹 Deleting vendor folder...
    rmdir /s /q vendor
)

REM Step 1: Generate Ent for shared
echo 👉 Generating Ent for shared
go run -mod=mod entgo.io/ent/cmd/ent generate ./shared/db/ent/schema --target ./shared/db/ent/generated --feature sql/execquery
IF ERRORLEVEL 1 GOTO error

REM Step 1.1: Init db
echo 👉 Initializing database
go run scripts/init_db/main.go
IF ERRORLEVEL 1 GOTO error

REM Step 2: Generate Ent for auditlog module
echo 👉 Generating Ent for auditlog
go run -mod=mod entgo.io/ent/cmd/ent generate ./modules/auditlog/ent/schema --target ./modules/auditlog/ent/generated
IF ERRORLEVEL 1 GOTO error

REM Step 2: Generate Ent for auditlog module
echo 👉 Generating Ent for push_notification
go run -mod=mod entgo.io/ent/cmd/ent generate ./modules/push_notification/ent/schema --target ./modules/push_notification/ent/generated
IF ERRORLEVEL 1 GOTO error

REM Step 3: Tidy & Vendor
echo 👉 Running go mod tidy
go mod tidy
IF ERRORLEVEL 1 GOTO error

echo 👉 Running go mod vendor
go mod vendor
IF ERRORLEVEL 1 GOTO error

@REM REM Step 4: Init roles
@REM echo 👉 Initializing roles
@REM go run scripts/init_roles/main.go
@REM IF ERRORLEVEL 1 GOTO error

REM Step 5: Build all
echo 👉 Building all modules
go build ./...
IF ERRORLEVEL 1 GOTO error

echo ✅ Project initialized successfully!
GOTO end

:error
echo ❌ Something went wrong.
exit /b 1

:end
