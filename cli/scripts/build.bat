@echo off
setlocal

set VERSION=%1
if "%VERSION%"=="" set VERSION=dev

for /f %%i in ('powershell -Command "Get-Date -Format yyyy-MM-ddTHH:mm:ssZ"') do set BUILD_DATE=%%i

set LDFLAGS=-X github.com/vknow360/otaship/cli/internal/commands.Version=%VERSION% -X github.com/vknow360/otaship/cli/internal/commands.BuildDate=%BUILD_DATE%

echo Building otaship-cli v%VERSION%

go build -ldflags "%LDFLAGS%" -o otaship-cli.exe cmd/otaship/main.go

if %ERRORLEVEL% equ 0 (
    echo Build successful
    otaship-cli.exe version
)