@echo off
setlocal EnableDelayedExpansion

set "VERSION_FILE=version.yaml"

REM ❌ Nếu version.yaml đã được staged, thì bỏ qua
git diff --cached --name-only | findstr /i "%VERSION_FILE%" >nul
if not errorlevel 1 (
  echo ⏭ %VERSION_FILE% already staged. Skipping bump.
  exit /b 0
)

REM 🧪 Đọc dòng version hiện tại
for /f "tokens=2 delims=:" %%A in ('findstr "^version:" %VERSION_FILE%') do (
  set "raw=%%A"
)

REM ⚙️ Tách base và build
for /f "tokens=1,2 delims=+" %%A in ("!raw!") do (
  set "base=%%A"
  set "build=%%B"
)

REM ⚠️ Strip space
set "base=!base: =!"
set "build=!build: =!"

if "!build!"=="" set build=0
set /a newbuild=!build! + 1

REM ✍️ Ghi lại version mới
> %VERSION_FILE% echo version: !base!+!newbuild!
git add %VERSION_FILE%

echo ✅ Bumped version to !base!+!newbuild! (on develop)
